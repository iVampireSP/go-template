package filter_test

import (
	"testing"

	"github.com/iVampireSP/go-template/pkg/filter"
)

func TestBuilder_Basic(t *testing.T) {
	// Test basic string filter
	filters := filter.New().
		String("status", "open").
		String("priority", "high").
		Build()

	if filters["status"] != "open" {
		t.Errorf("Expected status=open, got %v", filters["status"])
	}
	if filters["priority"] != "high" {
		t.Errorf("Expected priority=high, got %v", filters["priority"])
	}
}

func TestBuilder_Uint(t *testing.T) {
	// Test uint filter
	filters := filter.New().
		Uint("user_id", "123").
		Uint("assigned_to", "456").
		Build()

	if filters["user_id"] != uint(123) {
		t.Errorf("Expected user_id=123, got %v", filters["user_id"])
	}
	if filters["assigned_to"] != uint(456) {
		t.Errorf("Expected assigned_to=456, got %v", filters["assigned_to"])
	}
}

func TestBuilder_EmptyValues(t *testing.T) {
	// Test that empty values are not added
	filters := filter.New().
		String("status", "").
		Uint("user_id", "").
		Build()

	if len(filters) != 0 {
		t.Errorf("Expected empty filters, got %v", filters)
	}
}

func TestBuilder_InvalidUint(t *testing.T) {
	// Test that invalid int values are ignored
	filters := filter.New().
		Uint("user_id", "invalid").
		String("status", "open").
		Build()

	if _, exists := filters["user_id"]; exists {
		t.Error("Expected user_id to not exist due to invalid value")
	}
	if filters["status"] != "open" {
		t.Errorf("Expected status=open, got %v", filters["status"])
	}
}

func TestBuilder_Chaining(t *testing.T) {
	// Test method chaining
	builder := filter.New()
	result := builder.
		String("a", "1").
		String("b", "2").
		String("c", "3")

	if result != builder {
		t.Error("Expected chaining to return same builder instance")
	}

	filters := builder.Build()
	if len(filters) != 3 {
		t.Errorf("Expected 3 filters, got %d", len(filters))
	}
}

func TestBuilder_UintValue(t *testing.T) {
	// Test direct uint value
	filters := filter.New().
		UintValue("id", 100).
		UintValue("zero", 0). // Should not be added
		Build()

	if filters["id"] != uint(100) {
		t.Errorf("Expected id=100, got %v", filters["id"])
	}
	if _, exists := filters["zero"]; exists {
		t.Error("Expected zero value to not be added")
	}
}

func TestBuilder_Where(t *testing.T) {
	// Test generic Where method
	filters := filter.New().
		Where("custom", "value").
		Where("nil_value", nil). // Should not be added
		Build()

	if filters["custom"] != "value" {
		t.Errorf("Expected custom=value, got %v", filters["custom"])
	}
	if _, exists := filters["nil_value"]; exists {
		t.Error("Expected nil value to not be added")
	}
}

func TestBuilder_Has(t *testing.T) {
	builder := filter.New().
		String("status", "open")

	if !builder.Has("status") {
		t.Error("Expected Has to return true for existing filter")
	}
	if builder.Has("nonexistent") {
		t.Error("Expected Has to return false for non-existing filter")
	}
}

func TestBuilder_Count(t *testing.T) {
	builder := filter.New().
		String("a", "1").
		String("b", "2").
		String("c", "3")

	if builder.Count() != 3 {
		t.Errorf("Expected count=3, got %d", builder.Count())
	}
}

func TestBuilder_Remove(t *testing.T) {
	filters := filter.New().
		String("a", "1").
		String("b", "2").
		Remove("a").
		Build()

	if _, exists := filters["a"]; exists {
		t.Error("Expected 'a' to be removed")
	}
	if filters["b"] != "2" {
		t.Errorf("Expected b=2, got %v", filters["b"])
	}
}

func TestBuilder_Clear(t *testing.T) {
	builder := filter.New().
		String("a", "1").
		String("b", "2").
		Clear()

	if builder.Count() != 0 {
		t.Errorf("Expected count=0 after Clear, got %d", builder.Count())
	}
}

// Mock QueryGetter for testing
type mockQueryGetter struct {
	params map[string]string
}

func (m *mockQueryGetter) QueryParam(name string) string {
	return m.params[name]
}

func TestTicketFiltersFromQuery(t *testing.T) {
	mock := &mockQueryGetter{
		params: map[string]string{
			"user_id":     "100",
			"status":      "open",
			"priority":    "high",
			"department":  "support",
			"assigned_to": "200",
		},
	}

	filters := filter.TicketFiltersFromQuery(mock).Build()

	if filters["user_id"] != uint(100) {
		t.Errorf("Expected user_id=100, got %v", filters["user_id"])
	}
	if filters["status"] != "open" {
		t.Errorf("Expected status=open, got %v", filters["status"])
	}
	if filters["priority"] != "high" {
		t.Errorf("Expected priority=high, got %v", filters["priority"])
	}
	if filters["department"] != "support" {
		t.Errorf("Expected department=support, got %v", filters["department"])
	}
	if filters["assigned_to"] != uint(200) {
		t.Errorf("Expected assigned_to=200, got %v", filters["assigned_to"])
	}
}
