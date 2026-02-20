package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Store struct {
	db                *sql.DB
	idemTTL           time.Duration
	idemCleanupEvery  time.Duration
	cleanupDeletedSum int64
}

type IdempotencyRecord struct {
	StatusCode int
	Body       []byte
}

type App struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	Name      string         `json:"name"`
	Blueprint map[string]any `json:"blueprint"`
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func New(ctx context.Context, databaseURL string, idemTTLSeconds, idemCleanupSeconds int) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &Store{
		db:               db,
		idemTTL:          time.Duration(idemTTLSeconds) * time.Second,
		idemCleanupEvery: time.Duration(idemCleanupSeconds) * time.Second,
	}
	if s.idemTTL <= 0 {
		s.idemTTL = 24 * time.Hour
	}
	if s.idemCleanupEvery <= 0 {
		s.idemCleanupEvery = time.Minute
	}
	if err := s.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) StartIdempotencyCleanup(ctx context.Context) {
	t := time.NewTicker(s.idemCleanupEvery)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_, _ = s.CleanupExpiredIdempotency(ctx)
			}
		}
	}()
}

func (s *Store) CleanupExpiredIdempotency(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM idempotency_records WHERE expires_at <= NOW()`) //nolint:gosec
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	s.cleanupDeletedSum += n
	return n, nil
}

func (s *Store) IdempotencyCleanupDeletedTotal() int64 {
	return s.cleanupDeletedSum
}

func (s *Store) GetIdempotency(ctx context.Context, tenantID, endpoint, key string) (IdempotencyRecord, bool, error) {
	var status int
	var body []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT response_status, response_body
		FROM idempotency_records
		WHERE tenant_id = $1 AND endpoint = $2 AND idem_key = $3 AND expires_at > NOW()
	`, tenantID, endpoint, key).Scan(&status, &body)
	if errors.Is(err, sql.ErrNoRows) {
		return IdempotencyRecord{}, false, nil
	}
	if err != nil {
		return IdempotencyRecord{}, false, err
	}
	return IdempotencyRecord{StatusCode: status, Body: body}, true, nil
}

func (s *Store) PutIdempotency(ctx context.Context, tenantID, endpoint, key string, status int, body []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO idempotency_records (tenant_id, endpoint, idem_key, response_status, response_body, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW() + $6::interval)
		ON CONFLICT (tenant_id, endpoint, idem_key)
		DO UPDATE SET response_status = EXCLUDED.response_status, response_body = EXCLUDED.response_body, expires_at = EXCLUDED.expires_at
	`, tenantID, endpoint, key, status, body, fmt.Sprintf("%d seconds", int(s.idemTTL.Seconds())))
	return err
}

func (s *Store) SaveDecision(ctx context.Context, decisionID, tenantID, decisionHash, policyVersion, dataVersion string, generatedAt time.Time, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO decisions (decision_id, tenant_id, decision_hash, policy_version, data_version, generated_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (decision_id)
		DO UPDATE SET payload = EXCLUDED.payload, decision_hash = EXCLUDED.decision_hash, policy_version = EXCLUDED.policy_version, data_version = EXCLUDED.data_version, generated_at = EXCLUDED.generated_at
	`, decisionID, tenantID, decisionHash, policyVersion, dataVersion, generatedAt, payload)
	return err
}

func (s *Store) GetDecisionPayload(ctx context.Context, decisionID, tenantID string) ([]byte, bool, error) {
	var payload []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT payload
		FROM decisions
		WHERE decision_id = $1 AND tenant_id = $2
	`, decisionID, tenantID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return payload, true, nil
}

func (s *Store) CreateApp(ctx context.Context, app App) (App, error) {
	blueprint, err := json.Marshal(app.Blueprint)
	if err != nil {
		return App{}, err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO apps (id, tenant_id, name, blueprint, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, app.ID, app.TenantID, app.Name, blueprint, app.Version, app.CreatedAt, app.UpdatedAt)
	if err != nil {
		return App{}, err
	}
	return app, nil
}

func (s *Store) GetApp(ctx context.Context, tenantID, appID string) (App, bool, error) {
	var (
		name      string
		blueprint []byte
		version   int
		createdAt time.Time
		updatedAt time.Time
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT name, blueprint, version, created_at, updated_at
		FROM apps
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, appID).Scan(&name, &blueprint, &version, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return App{}, false, nil
	}
	if err != nil {
		return App{}, false, err
	}
	bp := map[string]any{}
	if err := json.Unmarshal(blueprint, &bp); err != nil {
		return App{}, false, err
	}
	return App{ID: appID, TenantID: tenantID, Name: name, Blueprint: bp, Version: version, CreatedAt: createdAt, UpdatedAt: updatedAt}, true, nil
}

func (s *Store) UpdateApp(ctx context.Context, app App) error {
	blueprint, err := json.Marshal(app.Blueprint)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE apps
		SET name = $3, blueprint = $4, version = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`, app.TenantID, app.ID, app.Name, blueprint, app.Version, app.UpdatedAt)
	return err
}

func (s *Store) SaveMutation(ctx context.Context, mutationID, tenantID, appID, class string, before, after, mutationPayload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO app_mutations (mutation_id, tenant_id, app_id, mutation_class, before_snapshot, after_snapshot, mutation_payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, mutationID, tenantID, appID, class, before, after, mutationPayload)
	return err
}

func (s *Store) SaveVerifyReport(ctx context.Context, reportID, tenantID, appID string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO verify_reports (report_id, tenant_id, app_id, payload, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, reportID, tenantID, appID, payload)
	return err
}

func (s *Store) SaveDeployIntent(ctx context.Context, intentID, tenantID, appID, target string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO deploy_intents (intent_id, tenant_id, app_id, target, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, intentID, tenantID, appID, target, payload)
	return err
}

func (s *Store) SaveStudioJob(ctx context.Context, tenantID, jobID string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO studio_jobs (job_id, tenant_id, payload, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, NOW(), NOW())
		ON CONFLICT (job_id)
		DO UPDATE SET payload = EXCLUDED.payload, updated_at = NOW()
	`, jobID, tenantID, string(payload))
	return err
}

func (s *Store) GetStudioJob(ctx context.Context, tenantID, jobID string) ([]byte, bool, error) {
	var payload []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT payload
		FROM studio_jobs
		WHERE tenant_id = $1 AND job_id = $2
	`, tenantID, jobID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return payload, true, nil
}

func (s *Store) initSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS decisions (
			decision_id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			decision_hash TEXT NOT NULL,
			policy_version TEXT NOT NULL,
			data_version TEXT NOT NULL,
			generated_at TIMESTAMPTZ NOT NULL,
			payload BYTEA NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS idempotency_records (
			tenant_id TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			idem_key TEXT NOT NULL,
			response_status INT NOT NULL,
			response_body BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			PRIMARY KEY (tenant_id, endpoint, idem_key)
		)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			blueprint JSONB NOT NULL,
			version INT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS app_mutations (
			mutation_id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			app_id TEXT NOT NULL,
			mutation_class TEXT NOT NULL,
			before_snapshot JSONB NOT NULL,
			after_snapshot JSONB NOT NULL,
			mutation_payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS verify_reports (
			report_id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			app_id TEXT NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS deploy_intents (
			intent_id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			app_id TEXT NOT NULL,
			target TEXT NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS studio_jobs (
			job_id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	// Keep byte-exact replay and idempotency payloads if prior schema used JSONB.
	migrations := []string{
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'decisions' AND column_name = 'payload' AND data_type = 'jsonb'
			) THEN
				ALTER TABLE decisions
				ALTER COLUMN payload TYPE BYTEA USING convert_to(payload::text, 'UTF8');
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'idempotency_records' AND column_name = 'response_body' AND data_type = 'jsonb'
			) THEN
				ALTER TABLE idempotency_records
				ALTER COLUMN response_body TYPE BYTEA USING convert_to(response_body::text, 'UTF8');
			END IF;
		END $$`,
	}
	for _, stmt := range migrations {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
