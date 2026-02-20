package config

import "os"

type Config struct {
	Port          string
	PolicyVersion string
	DataVersion   string
	DatabaseURL   string

	IdempotencyTTLSeconds     int
	IdempotencyCleanupSeconds int

	AuthTokens string

	GorseBaseURL string
	GorseAPIKey  string

	LLMDefaultProvider      string
	LLMDefaultModel         string
	LLMRequestTimeoutSecond int

	OllamaBaseURL      string
	OllamaDefaultModel string

	FrontierBaseURL      string
	FrontierAPIKey       string
	FrontierDefaultModel string
}

func Load() Config {
	return Config{
		Port:                      getenv("PORT", "4020"),
		PolicyVersion:             getenv("POLICY_VERSION", "policy-v1"),
		DataVersion:               getenv("DATA_VERSION", "data-v1"),
		DatabaseURL:               getenv("DATABASE_URL", "postgres://vda:vda@postgres:5432/vda?sslmode=disable"),
		IdempotencyTTLSeconds:     getenvInt("IDEMPOTENCY_TTL_SECONDS", 86400),
		IdempotencyCleanupSeconds: getenvInt("IDEMPOTENCY_CLEANUP_SECONDS", 60),
		AuthTokens:                getenv("AUTH_TOKENS", "dev-token:t_acme:dev-user"),
		GorseBaseURL:              getenv("GORSE_BASE_URL", "http://gorse:8088"),
		GorseAPIKey:               getenv("GORSE_API_KEY", "vda-demo-key"),
		LLMDefaultProvider:        getenv("LLM_DEFAULT_PROVIDER", "ollama"),
		LLMDefaultModel:           getenv("LLM_DEFAULT_MODEL", "glm-4.7-flash:latest"),
		LLMRequestTimeoutSecond:   getenvInt("LLM_REQUEST_TIMEOUT_SECONDS", 45),
		OllamaBaseURL:             getenv("OLLAMA_BASE_URL", "http://host.docker.internal:11434"),
		OllamaDefaultModel:        getenv("OLLAMA_DEFAULT_MODEL", "glm-4.7-flash:latest"),
		FrontierBaseURL:           getenv("FRONTIER_BASE_URL", "http://host.docker.internal:11434/v1"),
		FrontierAPIKey:            getenv("FRONTIER_API_KEY", ""),
		FrontierDefaultModel:      getenv("FRONTIER_DEFAULT_MODEL", "glm-4.7-flash:latest"),
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func getenvInt(k string, fallback int) int {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = (n * 10) + int(ch-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}
