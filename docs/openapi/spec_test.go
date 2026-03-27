package openapi

import (
	"encoding/json"
	"slices"
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

	orderSchema, ok := schemas["Order"].(map[string]any)
	if !ok {
		t.Fatal("expected Order schema")
	}

	orderProperties, ok := orderSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected Order properties")
	}

	statusProperty, ok := orderProperties["status"].(map[string]any)
	if !ok {
		t.Fatal("expected Order.status schema")
	}

	rawEnum, ok := statusProperty["enum"].([]any)
	if !ok {
		t.Fatal("expected Order.status enum")
	}

	enumValues := make([]string, 0, len(rawEnum))
	for _, value := range rawEnum {
		enumValue, ok := value.(string)
		if !ok {
			t.Fatalf("expected string enum value, got %#v", value)
		}

		enumValues = append(enumValues, enumValue)
	}

	for _, requiredStatus := range []string{"pending", "shipped", "cancelled"} {
		if !slices.Contains(enumValues, requiredStatus) {
			t.Fatalf("expected Order.status enum to include %q, got %#v", requiredStatus, enumValues)
		}
	}
}
