package filter

import (
	"strconv"
)

// Builder provides a fluent interface for building query filters
// Similar to Laravel's query builder pattern
type Builder struct {
	filters map[string]any
}

// New creates a new filter builder
func New() *Builder {
	return &Builder{
		filters: make(map[string]any),
	}
}

// String adds a string filter if the value is not empty
func (b *Builder) String(key, value string) *Builder {
	if value != "" {
		b.filters[key] = value
	}
	return b
}

// Uint adds a uint filter from a string value
// Returns the builder for chaining even if parsing fails
func (b *Builder) Uint(key, value string) *Builder {
	if value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 32); err == nil {
			b.filters[key] = uint(parsed)
		}
	}
	return b
}

// UintValue adds a uint filter directly
func (b *Builder) UintValue(key string, value uint) *Builder {
	if value > 0 {
		b.filters[key] = value
	}
	return b
}

// Int adds an int filter from a string value
func (b *Builder) Int(key, value string) *Builder {
	if value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			b.filters[key] = parsed
		}
	}
	return b
}

// IntValue adds an int filter directly
func (b *Builder) IntValue(key string, value int) *Builder {
	b.filters[key] = value
	return b
}

// Bool adds a bool filter from a string value
func (b *Builder) Bool(key, value string) *Builder {
	if value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			b.filters[key] = parsed
		}
	}
	return b
}

// BoolValue adds a bool filter directly
func (b *Builder) BoolValue(key string, value bool) *Builder {
	b.filters[key] = value
	return b
}

// Where adds a filter with any value type
func (b *Builder) Where(key string, value any) *Builder {
	if value != nil {
		b.filters[key] = value
	}
	return b
}

// WhereIn adds an IN filter (for slice values)
func (b *Builder) WhereIn(key string, values []any) *Builder {
	if len(values) > 0 {
		b.filters[key] = values
	}
	return b
}

// Build returns the built filters map
func (b *Builder) Build() map[string]any {
	return b.filters
}

// Get returns a specific filter value
func (b *Builder) Get(key string) (any, bool) {
	val, exists := b.filters[key]
	return val, exists
}

// Has checks if a filter exists
func (b *Builder) Has(key string) bool {
	_, exists := b.filters[key]
	return exists
}

// Count returns the number of filters
func (b *Builder) Count() int {
	return len(b.filters)
}

// Clear removes all filters
func (b *Builder) Clear() *Builder {
	b.filters = make(map[string]any)
	return b
}

// Remove removes a specific filter
func (b *Builder) Remove(key string) *Builder {
	delete(b.filters, key)
	return b
}
