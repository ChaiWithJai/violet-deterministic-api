package main

import (
	"log"
	"net/http"

	"generated/bundle-app/services/api/internal/runtime"
	"generated/bundle-app/services/api/internal/tools"
)

func main() {
	spec := runtime.Spec{
		AppName:    "Bundle App",
		Domain:     "saas",
		Plan:       "starter",
		Region:     "us-east-1",
		Users:      []string{"admin", "operator"},
		Entities:   []string{"account", "workspace", "activity"},
		Workflows:  []string{"create_record", "approve_record", "notify_user"},
		ToolRoutes: tools.Catalog(),
	}

	server := runtime.NewServer(spec)
	log.Println("listening on :8090")
	if err := http.ListenAndServe(":8090", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
