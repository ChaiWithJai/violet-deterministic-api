package main

import (
	"log"
	"net/http"

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
	log.Println("listening on :8090")
	if err := http.ListenAndServe(":8090", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
