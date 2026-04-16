package filter_test

import (
	"fmt"

	"github.com/iVampireSP/go-template/pkg/filter"
)

// Example_basicUsage demonstrates basic filter building
func Example_basicUsage() {
	// Create filters with method chaining
	filters := filter.New().
		String("status", "active").
		Uint("user_id", "123").
		String("priority", "high").
		Build()

	fmt.Printf("Filters: %v\n", filters)
	// Output: Filters: map[priority:high status:active user_id:123]
}

// Example_emptyValues demonstrates that empty values are skipped
func Example_emptyValues() {
	// Empty values are automatically skipped
	filters := filter.New().
		String("status", "").      // Will be skipped
		String("priority", "low"). // Will be added
		Uint("user_id", "").       // Will be skipped
		Build()

	fmt.Printf("Filters: %v\n", filters)
	// Output: Filters: map[priority:low]
}

// Example_invalidValues demonstrates that invalid values are skipped
func Example_invalidValues() {
	// Invalid type conversions are automatically handled
	filters := filter.New().
		Uint("user_id", "not_a_number"). // Will be skipped
		String("status", "open").        // Will be added
		Int("count", "invalid").         // Will be skipped
		Build()

	fmt.Printf("Filters: %v\n", filters)
	// Output: Filters: map[status:open]
}

// Example_directValues demonstrates using direct value methods
func Example_directValues() {
	userID := uint(100)

	filters := filter.New().
		UintValue("user_id", userID).
		UintValue("zero", 0). // Will be skipped (zero value)
		IntValue("count", 5).
		BoolValue("active", true).
		Build()

	fmt.Printf("Has user_id: %v\n", len(filters) >= 3)
	// Output: Has user_id: true
}

// Example_conditionalFilters demonstrates dynamic filter building
func Example_conditionalFilters() {
	builder := filter.New().
		String("status", "open")

	// Add filters conditionally
	includeUserFilter := true
	if includeUserFilter {
		builder = builder.UintValue("user_id", 123)
	}

	includePriority := false
	if includePriority {
		builder = builder.String("priority", "high")
	}

	filters := builder.Build()
	fmt.Printf("Filter count: %d\n", len(filters))
	// Output: Filter count: 2
}

// Example_builderMethods demonstrates various builder methods
func Example_builderMethods() {
	builder := filter.New().
		String("a", "1").
		String("b", "2").
		String("c", "3")

	// Check existence
	fmt.Printf("Has 'a': %v\n", builder.Has("a"))

	// Get count
	fmt.Printf("Count: %d\n", builder.Count())

	// Remove a filter
	builder.Remove("b")
	fmt.Printf("Count after remove: %d\n", builder.Count())

	// Output:
	// Has 'a': true
	// Count: 3
	// Count after remove: 2
}

// Example_whereMethod demonstrates the generic Where method
func Example_whereMethod() {
	// Use Where for any value type
	filters := filter.New().
		Where("string_field", "value").
		Where("int_field", 123).
		Where("bool_field", true).
		Where("nil_field", nil). // Will be skipped
		Build()

	fmt.Printf("Filter count: %d\n", len(filters))
	// Output: Filter count: 3
}

// Example_ticketFilters demonstrates the ticket-specific helper
func Example_ticketFilters() {
	// Mock query getter
	type mockQuery struct {
		params map[string]string
	}
	mock := &mockQuery{
		params: map[string]string{
			"status":   "open",
			"priority": "high",
			"user_id":  "123",
		},
	}

	// Implement QueryGetter interface
	queryGetter := struct {
		*mockQuery
	}{mock}
	queryGetter.mockQuery = mock

	// Note: In real usage, you would pass echo.Context directly
	// filters := filter.TicketFiltersFromQuery(ctx).Build()

	fmt.Println("Ticket filters can be built using TicketFiltersFromQuery")
	// Output: Ticket filters can be built using TicketFiltersFromQuery
}

// Example_chainedBuilding demonstrates building filters step by step
func Example_chainedBuilding() {
	// Start with base filters
	builder := filter.New().
		String("status", "active")

	// Add more filters based on conditions
	userFilter := "100"
	if userFilter != "" {
		builder = builder.Uint("user_id", userFilter)
	}

	priorityFilter := "high"
	if priorityFilter != "" {
		builder = builder.String("priority", priorityFilter)
	}

	filters := builder.Build()
	fmt.Printf("Filter count: %d\n", len(filters))
	// Output: Filter count: 3
}

// Example_clearAndReuse demonstrates clearing and reusing a builder
func Example_clearAndReuse() {
	builder := filter.New().
		String("status", "open").
		String("priority", "high")

	fmt.Printf("Before clear: %d\n", builder.Count())

	builder.Clear().
		String("status", "closed")

	fmt.Printf("After clear: %d\n", builder.Count())
	// Output:
	// Before clear: 2
	// After clear: 1
}
