package osa

import (
	"context"
	"net/http"
)

type RuntimeModel struct {
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	ContextWindow int    `json:"context_window"`
}
type RuntimeProvider struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Configured   bool   `json:"configured"`
	Connected    bool   `json:"connected"`
	Type         string `json:"type"`
	DefaultModel string `json:"default_model"`
}
type ModelCatalog struct {
	Provider  string            `json:"provider"`
	Current   string            `json:"current"`
	Models    []RuntimeModel    `json:"models"`
	Providers []RuntimeProvider `json:"providers"`
}

// ModelCatalog reads the installed runtime, never a separate local model server.
func (r *ResilientClient) ModelCatalog(ctx context.Context) (*ModelCatalog, error) {
	var catalog ModelCatalog
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/v1/models", nil, &catalog); err != nil {
		return nil, err
	}
	var providers struct {
		Providers []RuntimeProvider `json:"providers"`
	}
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/v1/providers", nil, &providers); err != nil {
		return nil, err
	}
	var current HealthResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/v1/models/current", nil, &current); err != nil {
		return nil, err
	}
	catalog.Providers = providers.Providers
	catalog.Provider = current.Provider
	catalog.Current = current.Model
	// The effective model can include a tag that is absent from the catalog.
	found := false
	for _, model := range catalog.Models {
		if model.Provider == current.Provider && model.Name == current.Model {
			found = true
			break
		}
	}
	if !found && current.Model != "" {
		catalog.Models = append([]RuntimeModel{{Name: current.Model, Provider: current.Provider}}, catalog.Models...)
	}
	if catalog.Models == nil {
		catalog.Models = []RuntimeModel{}
	}
	if catalog.Providers == nil {
		catalog.Providers = []RuntimeProvider{}
	}
	return &catalog, nil
}
