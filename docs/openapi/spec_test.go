package openapi

import (
	"encoding/json"
	"testing"
)

func TestSpecDocumentsProductsAndOrdersSurface(t *testing.T) {
	t.Parallel()

	var document map[string]any
	if err := json.Unmarshal(Spec(), &document); err != nil {
		t.Fatalf("decode spec: %v", err)
	}

	if got := document["openapi"]; got != "3.0.3" {
		t.Fatalf("expected openapi version %q, got %#v", "3.0.3", got)
	}

	paths, ok := document["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected paths object")
	}

	requiredPaths := []string{
		"/v1/products",
		"/v1/products/{id}",
		"/v1/search/products",
		"/v1/orders",
		"/v1/orders/{id}",
		"/v1/orders/{id}/cancel",
	}

	for _, path := range requiredPaths {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected path %q to be documented", path)
		}
	}

	components, ok := document["components"].(map[string]any)
	if !ok {
		t.Fatal("expected components object")
	}

	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		t.Fatal("expected component schemas")
	}

	createRequest, ok := schemas["CreateProductRequest"].(map[string]any)
	if !ok {
		t.Fatal("expected CreateProductRequest schema")
	}

	properties, ok := createRequest["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected CreateProductRequest properties")
	}

	sku, ok := properties["sku"].(map[string]any)
	if !ok {
		t.Fatal("expected deprecated sku alias in create request schema")
	}

	if deprecated, ok := sku["deprecated"].(bool); !ok || !deprecated {
		t.Fatalf("expected sku to be marked deprecated, got %#v", sku["deprecated"])
	}
}
