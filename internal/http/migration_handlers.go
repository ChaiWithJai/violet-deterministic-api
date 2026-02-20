package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	httpstd "net/http"
	"sort"
	"strings"
	"time"

	"github.com/restarone/violet-deterministic-api/internal/storage"
)

const defaultVioletBundleVersion = "violet-export-v1"

type violetExportRequest struct {
	AppID         string         `json:"app_id,omitempty"`
	Namespace     string         `json:"namespace,omitempty"`
	Source        map[string]any `json:"source,omitempty"`
	BundleVersion string         `json:"bundle_version,omitempty"`
}

type violetImportRequest struct {
	AppID        string       `json:"app_id,omitempty"`
	AllowPartial bool         `json:"allow_partial,omitempty"`
	Bundle       violetBundle `json:"bundle"`
}

type violetBundle struct {
	BundleID          string           `json:"bundle_id"`
	SourceSystem      string           `json:"source_system"`
	BundleVersion     string           `json:"bundle_version"`
	PolicyVersion     string           `json:"policy_version"`
	DataVersion       string           `json:"data_version"`
	Namespace         string           `json:"namespace"`
	Resources         []violetResource `json:"resources"`
	Actions           []violetAction   `json:"actions,omitempty"`
	Roles             []string         `json:"roles,omitempty"`
	UnsupportedFields []string         `json:"unsupported_fields,omitempty"`
	Checksum          string           `json:"checksum"`
}

type violetResource struct {
	Name    string            `json:"name"`
	Fields  map[string]string `json:"fields,omitempty"`
	Records []map[string]any  `json:"records,omitempty"`
}

type violetAction struct {
	Name     string         `json:"name"`
	Resource string         `json:"resource,omitempty"`
	Type     string         `json:"type,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
}

func (s *Server) handleMigrationExport(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}

	var req violetExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}

	var (
		source map[string]any
		found  bool
		err    error
	)
	if strings.TrimSpace(req.AppID) != "" {
		source, found, err = s.exportSourceFromApp(r.Context(), claims.TenantID, strings.TrimSpace(req.AppID))
		if err != nil {
			writeError(w, httpstd.StatusInternalServerError, "app_read_failed", map[string]any{"details": err.Error()})
			return
		}
		if !found {
			writeError(w, httpstd.StatusNotFound, "app_not_found", nil)
			return
		}
	} else {
		source = copyMap(req.Source)
	}

	if source == nil {
		source = map[string]any{}
	}
	if strings.TrimSpace(req.Namespace) != "" {
		source["namespace"] = strings.TrimSpace(req.Namespace)
	}

	bundleVersion := strings.TrimSpace(req.BundleVersion)
	if bundleVersion == "" {
		bundleVersion = readString(source, "bundle_version")
	}
	if bundleVersion == "" {
		bundleVersion = defaultVioletBundleVersion
	}

	policyVersion := readString(source, "policy_version")
	if policyVersion == "" {
		policyVersion = s.cfg.PolicyVersion
	}
	dataVersion := readString(source, "data_version")
	if dataVersion == "" {
		dataVersion = s.cfg.DataVersion
	}

	bundle, err := buildVioletBundle(source, claims.TenantID, bundleVersion, policyVersion, dataVersion)
	if err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_export_request", map[string]any{"details": err.Error()})
		return
	}

	resp := map[string]any{
		"bundle": bundle,
		"counts": map[string]int{
			"resources": len(bundle.Resources),
			"actions":   len(bundle.Actions),
			"roles":     len(bundle.Roles),
		},
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		writeError(w, httpstd.StatusInternalServerError, "marshal_failed", nil)
		return
	}
	if err := s.store.SaveMigrationBundle(r.Context(), bundle.BundleID, claims.TenantID, "export", payload); err != nil {
		writeError(w, httpstd.StatusInternalServerError, "migration_bundle_write_failed", map[string]any{"details": err.Error()})
		return
	}
	writeJSON(w, httpstd.StatusOK, payload)
}

func (s *Server) handleMigrationImport(w httpstd.ResponseWriter, r *httpstd.Request) {
	claims, ok := s.authClaims(w, r)
	if !ok {
		return
	}
	idemKey, ok := s.idempotencyKey(w, r)
	if !ok {
		return
	}

	var req violetImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_json", nil)
		return
	}
	if req.AllowPartial {
		writeError(w, httpstd.StatusBadRequest, "partial_apply_disabled", nil)
		return
	}

	bundle, err := normalizeImportedBundle(req.Bundle, claims.TenantID, s.cfg.PolicyVersion, s.cfg.DataVersion)
	if err != nil {
		writeError(w, httpstd.StatusBadRequest, "invalid_bundle", map[string]any{"details": err.Error()})
		return
	}

	s.withIdempotency(r.Context(), w, claims.TenantID, r.URL.Path, idemKey, func() (int, []byte, error) {
		now := time.Now().UTC()
		status := httpstd.StatusOK
		appID := strings.TrimSpace(req.AppID)

		var app storage.App
		if appID != "" {
			existing, found, err := s.store.GetApp(r.Context(), claims.TenantID, appID)
			if err != nil {
				return 0, nil, err
			}
			if !found {
				return httpstd.StatusNotFound, mustJSON(map[string]any{"error": "app_not_found"}), nil
			}
			app = existing
			app.Version++
			app.UpdatedAt = now
			if app.Blueprint == nil {
				app.Blueprint = map[string]any{}
			}
			if err := applyImportedBundle(&app, bundle); err != nil {
				return 0, nil, err
			}
			if err := s.store.UpdateApp(r.Context(), app); err != nil {
				return 0, nil, err
			}
		} else {
			status = httpstd.StatusCreated
			app = storage.App{
				ID:        stableID("app", claims.TenantID, bundle.Checksum, idemKey),
				TenantID:  claims.TenantID,
				Name:      bundle.Namespace,
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
				Blueprint: map[string]any{},
			}
			if err := applyImportedBundle(&app, bundle); err != nil {
				return 0, nil, err
			}
			created, err := s.store.CreateApp(r.Context(), app)
			if err != nil {
				return 0, nil, err
			}
			app = created
		}

		resp := map[string]any{
			"status":             "imported",
			"app":                app,
			"bundle_id":          bundle.BundleID,
			"checksum":           bundle.Checksum,
			"unsupported_fields": bundle.UnsupportedFields,
			"imported_counts": map[string]int{
				"resources": len(bundle.Resources),
				"actions":   len(bundle.Actions),
				"roles":     len(bundle.Roles),
			},
		}
		payload, err := json.Marshal(resp)
		if err != nil {
			return 0, nil, err
		}
		if err := s.store.SaveMigrationBundle(r.Context(), bundle.BundleID, claims.TenantID, "import", payload); err != nil {
			return 0, nil, err
		}
		return status, payload, nil
	})
}

func (s *Server) exportSourceFromApp(ctx context.Context, tenantID, appID string) (map[string]any, bool, error) {
	app, found, err := s.store.GetApp(ctx, tenantID, appID)
	if err != nil || !found {
		return nil, found, err
	}
	if app.Blueprint == nil {
		return map[string]any{"namespace": app.Name}, true, nil
	}
	if rawBundle, ok := app.Blueprint["migration_violet_bundle"].(map[string]any); ok {
		return copyMap(rawBundle), true, nil
	}

	source := map[string]any{}
	if v := readString(app.Blueprint, "namespace"); v != "" {
		source["namespace"] = v
	}
	if resources, ok := app.Blueprint["resources"].([]any); ok {
		source["resources"] = resources
	}
	if actions, ok := app.Blueprint["actions"].([]any); ok {
		source["actions"] = actions
	}
	if roles, ok := app.Blueprint["roles"].([]any); ok {
		source["roles"] = roles
	}
	if len(source) == 0 {
		source["namespace"] = app.Name
	}
	return source, true, nil
}

func applyImportedBundle(app *storage.App, bundle violetBundle) error {
	if app.Blueprint == nil {
		app.Blueprint = map[string]any{}
	}
	app.Name = bundle.Namespace
	app.Blueprint["namespace"] = bundle.Namespace
	app.Blueprint["resources"] = bundle.Resources
	app.Blueprint["actions"] = bundle.Actions
	app.Blueprint["roles"] = bundle.Roles
	app.Blueprint["migration_violet_bundle"] = map[string]any{
		"bundle_id":          bundle.BundleID,
		"source_system":      bundle.SourceSystem,
		"bundle_version":     bundle.BundleVersion,
		"policy_version":     bundle.PolicyVersion,
		"data_version":       bundle.DataVersion,
		"namespace":          bundle.Namespace,
		"resources":          bundle.Resources,
		"actions":            bundle.Actions,
		"roles":              bundle.Roles,
		"unsupported_fields": bundle.UnsupportedFields,
		"checksum":           bundle.Checksum,
	}
	return nil
}

func normalizeImportedBundle(in violetBundle, tenantID, defaultPolicyVersion, defaultDataVersion string) (violetBundle, error) {
	if strings.TrimSpace(in.SourceSystem) != "" && strings.TrimSpace(in.SourceSystem) != "violet-rails" {
		return violetBundle{}, fmt.Errorf("source_system must be violet-rails")
	}

	bundleVersion := strings.TrimSpace(in.BundleVersion)
	if bundleVersion == "" {
		bundleVersion = defaultVioletBundleVersion
	}
	policyVersion := strings.TrimSpace(in.PolicyVersion)
	if policyVersion == "" {
		policyVersion = defaultPolicyVersion
	}
	dataVersion := strings.TrimSpace(in.DataVersion)
	if dataVersion == "" {
		dataVersion = defaultDataVersion
	}

	source := map[string]any{
		"namespace":          in.Namespace,
		"resources":          resourcesToAny(in.Resources),
		"actions":            actionsToAny(in.Actions),
		"roles":              stringsToAny(in.Roles),
		"unsupported_fields": stringsToAny(in.UnsupportedFields),
	}
	out, err := buildVioletBundle(source, tenantID, bundleVersion, policyVersion, dataVersion)
	if err != nil {
		return violetBundle{}, err
	}
	if expected := strings.TrimSpace(in.Checksum); expected != "" && expected != out.Checksum {
		return violetBundle{}, fmt.Errorf("checksum mismatch: expected=%s got=%s", expected, out.Checksum)
	}
	return out, nil
}

func buildVioletBundle(source map[string]any, tenantID, bundleVersion, policyVersion, dataVersion string) (violetBundle, error) {
	namespace, resources, actions, roles, unsupported, err := normalizeVioletSource(source)
	if err != nil {
		return violetBundle{}, err
	}
	if namespace == "" {
		return violetBundle{}, fmt.Errorf("namespace is required")
	}

	out := violetBundle{
		SourceSystem:      "violet-rails",
		BundleVersion:     bundleVersion,
		PolicyVersion:     policyVersion,
		DataVersion:       dataVersion,
		Namespace:         namespace,
		Resources:         resources,
		Actions:           actions,
		Roles:             roles,
		UnsupportedFields: unsupported,
	}
	out.Checksum = hashVioletBundle(out)
	out.BundleID = stableID("mig", tenantID, out.Checksum)
	return out, nil
}

func normalizeVioletSource(source map[string]any) (string, []violetResource, []violetAction, []string, []string, error) {
	if source == nil {
		source = map[string]any{}
	}

	unsupported := []string{}
	namespace := strings.TrimSpace(readString(source, "namespace"))
	if namespace == "" {
		namespace = strings.TrimSpace(readString(source, "api_namespace"))
	}

	rawUnsupported, ok := source["unsupported_fields"]
	if ok {
		items, ok := rawUnsupported.([]any)
		if !ok {
			unsupported = append(unsupported, "unsupported_fields")
		} else {
			for i, item := range items {
				v, ok := item.(string)
				if !ok {
					unsupported = append(unsupported, fmt.Sprintf("unsupported_fields[%d]", i))
					continue
				}
				if strings.TrimSpace(v) != "" {
					unsupported = append(unsupported, strings.TrimSpace(v))
				}
			}
		}
	}

	resources, resourceUnsupported, err := normalizeResources(source["resources"])
	if err != nil {
		return "", nil, nil, nil, nil, err
	}
	unsupported = append(unsupported, resourceUnsupported...)

	actions, actionUnsupported, err := normalizeActions(source["actions"])
	if err != nil {
		return "", nil, nil, nil, nil, err
	}
	unsupported = append(unsupported, actionUnsupported...)

	roles, roleUnsupported, err := normalizeRoles(source["roles"])
	if err != nil {
		return "", nil, nil, nil, nil, err
	}
	unsupported = append(unsupported, roleUnsupported...)

	knownTopLevel := map[string]struct{}{
		"namespace":          {},
		"api_namespace":      {},
		"resources":          {},
		"actions":            {},
		"roles":              {},
		"unsupported_fields": {},
		"source_system":      {},
		"bundle_version":     {},
		"policy_version":     {},
		"data_version":       {},
		"bundle_id":          {},
		"checksum":           {},
	}
	for key := range source {
		if _, known := knownTopLevel[key]; !known {
			unsupported = append(unsupported, key)
		}
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Name != actions[j].Name {
			return actions[i].Name < actions[j].Name
		}
		if actions[i].Resource != actions[j].Resource {
			return actions[i].Resource < actions[j].Resource
		}
		return actions[i].Type < actions[j].Type
	})
	sort.Strings(roles)

	return namespace, resources, actions, uniqueSortedStrings(roles), uniqueSortedStrings(unsupported), nil
}

func normalizeResources(raw any) ([]violetResource, []string, error) {
	if raw == nil {
		return nil, nil, nil
	}
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []map[string]any:
		items = make([]any, 0, len(v))
		for _, item := range v {
			items = append(items, item)
		}
	case []violetResource:
		items = resourcesToAny(v)
	default:
		return nil, nil, fmt.Errorf("resources must be an array")
	}

	resources := make([]violetResource, 0, len(items))
	unsupported := []string{}
	for i, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			unsupported = append(unsupported, fmt.Sprintf("resources[%d]", i))
			continue
		}
		resource := violetResource{
			Fields:  map[string]string{},
			Records: []map[string]any{},
		}
		for key, value := range obj {
			switch key {
			case "name":
				v, ok := value.(string)
				if !ok || strings.TrimSpace(v) == "" {
					return nil, nil, fmt.Errorf("resources[%d].name must be non-empty string", i)
				}
				resource.Name = strings.TrimSpace(v)
			case "fields":
				switch fieldMap := value.(type) {
				case map[string]any:
					for fieldName, fieldType := range fieldMap {
						ft, ok := fieldType.(string)
						if !ok || strings.TrimSpace(ft) == "" {
							unsupported = append(unsupported, fmt.Sprintf("resources[%d].fields.%s", i, fieldName))
							continue
						}
						resource.Fields[fieldName] = strings.TrimSpace(ft)
					}
				case map[string]string:
					for fieldName, fieldType := range fieldMap {
						if strings.TrimSpace(fieldType) == "" {
							unsupported = append(unsupported, fmt.Sprintf("resources[%d].fields.%s", i, fieldName))
							continue
						}
						resource.Fields[fieldName] = strings.TrimSpace(fieldType)
					}
				default:
					unsupported = append(unsupported, fmt.Sprintf("resources[%d].fields", i))
				}
			case "records":
				switch records := value.(type) {
				case []any:
					for j, record := range records {
						recordMap, ok := record.(map[string]any)
						if !ok {
							unsupported = append(unsupported, fmt.Sprintf("resources[%d].records[%d]", i, j))
							continue
						}
						resource.Records = append(resource.Records, recordMap)
					}
				case []map[string]any:
					for _, recordMap := range records {
						resource.Records = append(resource.Records, recordMap)
					}
				default:
					unsupported = append(unsupported, fmt.Sprintf("resources[%d].records", i))
				}
			default:
				unsupported = append(unsupported, fmt.Sprintf("resources[%d].%s", i, key))
			}
		}
		if resource.Name == "" {
			return nil, nil, fmt.Errorf("resources[%d].name is required", i)
		}
		if len(resource.Fields) == 0 {
			resource.Fields = nil
		}
		if len(resource.Records) == 0 {
			resource.Records = nil
		}
		resources = append(resources, resource)
	}
	return resources, unsupported, nil
}

func normalizeActions(raw any) ([]violetAction, []string, error) {
	if raw == nil {
		return nil, nil, nil
	}
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []map[string]any:
		items = make([]any, 0, len(v))
		for _, item := range v {
			items = append(items, item)
		}
	case []violetAction:
		items = actionsToAny(v)
	default:
		return nil, nil, fmt.Errorf("actions must be an array")
	}

	actions := make([]violetAction, 0, len(items))
	unsupported := []string{}
	for i, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			unsupported = append(unsupported, fmt.Sprintf("actions[%d]", i))
			continue
		}
		action := violetAction{}
		for key, value := range obj {
			switch key {
			case "name":
				v, ok := value.(string)
				if !ok || strings.TrimSpace(v) == "" {
					return nil, nil, fmt.Errorf("actions[%d].name must be non-empty string", i)
				}
				action.Name = strings.TrimSpace(v)
			case "resource":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					action.Resource = strings.TrimSpace(v)
				}
			case "type":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					action.Type = strings.TrimSpace(v)
				}
			case "config":
				switch cfg := value.(type) {
				case map[string]any:
					action.Config = cfg
				case map[string]string:
					action.Config = map[string]any{}
					for k, v := range cfg {
						action.Config[k] = v
					}
				default:
					unsupported = append(unsupported, fmt.Sprintf("actions[%d].config", i))
				}
			default:
				unsupported = append(unsupported, fmt.Sprintf("actions[%d].%s", i, key))
			}
		}
		if action.Name == "" {
			return nil, nil, fmt.Errorf("actions[%d].name is required", i)
		}
		actions = append(actions, action)
	}
	return actions, unsupported, nil
}

func normalizeRoles(raw any) ([]string, []string, error) {
	if raw == nil {
		return nil, nil, nil
	}
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []string:
		items = stringsToAny(v)
	default:
		return nil, nil, fmt.Errorf("roles must be an array")
	}
	roles := make([]string, 0, len(items))
	unsupported := []string{}
	for i, item := range items {
		role, ok := item.(string)
		if !ok || strings.TrimSpace(role) == "" {
			unsupported = append(unsupported, fmt.Sprintf("roles[%d]", i))
			continue
		}
		roles = append(roles, strings.TrimSpace(role))
	}
	return roles, unsupported, nil
}

func hashVioletBundle(bundle violetBundle) string {
	core := map[string]any{
		"source_system":      bundle.SourceSystem,
		"bundle_version":     bundle.BundleVersion,
		"policy_version":     bundle.PolicyVersion,
		"data_version":       bundle.DataVersion,
		"namespace":          bundle.Namespace,
		"resources":          bundle.Resources,
		"actions":            bundle.Actions,
		"roles":              bundle.Roles,
		"unsupported_fields": bundle.UnsupportedFields,
	}
	raw, _ := json.Marshal(core)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func uniqueSortedStrings(in []string) []string {
	set := map[string]struct{}{}
	out := []string{}
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := set[item]; exists {
			continue
		}
		set[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func copyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := map[string]any{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func readString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok {
		return ""
	}
	v, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func resourcesToAny(in []violetResource) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		obj := map[string]any{
			"name": item.Name,
		}
		if len(item.Fields) > 0 {
			obj["fields"] = item.Fields
		}
		if len(item.Records) > 0 {
			obj["records"] = item.Records
		}
		out = append(out, obj)
	}
	return out
}

func actionsToAny(in []violetAction) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		obj := map[string]any{
			"name": item.Name,
		}
		if strings.TrimSpace(item.Resource) != "" {
			obj["resource"] = item.Resource
		}
		if strings.TrimSpace(item.Type) != "" {
			obj["type"] = item.Type
		}
		if len(item.Config) > 0 {
			obj["config"] = item.Config
		}
		out = append(out, obj)
	}
	return out
}

func stringsToAny(in []string) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		out = append(out, item)
	}
	return out
}
