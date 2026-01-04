package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestProductsHandler_Facets_MissingTenantID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, _, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	resp := makeRequest(t, server, "GET", "/api/v1/products/facets?type=goods")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	errResp := decodeErrorResponse(t, resp)
	if errResp.Error.Code != "VALIDATION_FAILED" {
		t.Errorf("Expected error code VALIDATION_FAILED, got %s", errResp.Error.Code)
	}
	if errResp.Error.Message == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestProductsHandler_Facets_WithTenantID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, _, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	resp := makeRequest(t, server, "GET", "/api/v1/products/facets?type=goods&tenant_id=test-tenant")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var payload FacetSchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if payload.TenantID != "test-tenant" {
		t.Errorf("Expected tenant_id test-tenant, got %s", payload.TenantID)
	}
	if payload.Domain != "goods" {
		t.Errorf("Expected domain goods, got %s", payload.Domain)
	}
	if len(payload.Facets) == 0 {
		t.Error("Expected facets schema to be non-empty")
	}
}
