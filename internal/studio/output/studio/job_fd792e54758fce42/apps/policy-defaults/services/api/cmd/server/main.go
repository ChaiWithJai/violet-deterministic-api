package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"generated/policy-defaults/services/api/internal/runtime"
	"generated/policy-defaults/services/api/internal/tools"
)

func main() {
	spec := runtime.Spec{
		AppName:    "Policy Defaults",
		Domain:     "saas",
		Plan:       "starter",
		Region:     "us-east-1",
		Users:      []string{"admin", "operator"},
		Entities:   []string{"account", "workspace", "activity"},
		Workflows:  []string{"create_record", "approve_record", "notify_user"},
		ToolRoutes: tools.Catalog(),
	}

	server := runtime.NewServer(spec)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	addr := ":" + port
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(fmt.Errorf("listen %s: %w", addr, err))
	}
}
