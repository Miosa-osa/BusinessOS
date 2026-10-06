package osa

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestModelCatalogUsesRuntimeAndRetainsCurrentTaggedModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/models":
			fmt.Fprint(w, `{"models":[{"name":"new-model","provider":"ollama_cloud"}]}`)
		case "/api/v1/providers":
			fmt.Fprint(w, `{"providers":[{"slug":"ollama_cloud","name":"Ollama Cloud","configured":true}]}`)
		case "/api/v1/models/current":
			fmt.Fprint(w, `{"model":"current:cloud","provider":"ollama"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewClient(&Config{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := (&ResilientClient{client: client}).ModelCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Current != "current:cloud" || len(catalog.Models) != 2 || catalog.Models[0].Name != "current:cloud" || catalog.Providers[0].Slug != "ollama_cloud" {
		t.Fatalf("unexpected catalog %#v", catalog)
	}
}
