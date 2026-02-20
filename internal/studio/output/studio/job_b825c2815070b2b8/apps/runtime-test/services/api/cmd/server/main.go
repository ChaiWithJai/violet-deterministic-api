package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"generated/runtime-test/services/api/internal/runtime"
	"generated/runtime-test/services/api/internal/tools"
)

func main() {
	spec := runtime.Spec{
		AppName:    "Runtime Test",
		Domain:     "crm",
		Plan:       "enterprise",
		Region:     "us-east-1",
		Users:      []string{"admin"},
		Entities:   []string{"account"},
		Workflows:  []string{"create_customer"},
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
