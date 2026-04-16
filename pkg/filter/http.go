package filter

// QueryGetter is an interface for getting query parameters
// This abstraction allows us to use filter.Builder without depending on specific HTTP frameworks
type QueryGetter interface {
	QueryParam(name string) string
}

// FromQuery creates a new Builder and provides helper methods for common query parameter extraction
// This helper function makes it easy to build filters from HTTP query parameters
func FromQuery(q QueryGetter) *Builder {
	return New()
}

// TicketFiltersFromQuery is a specialized helper for building ticket filters
// Usage: filters := filter.TicketFiltersFromQuery(ctx).Build()
func TicketFiltersFromQuery(q QueryGetter) *Builder {
	return New().
		Uint("user_id", q.QueryParam("user_id")).
		String("status", q.QueryParam("status")).
		String("priority", q.QueryParam("priority")).
		String("department", q.QueryParam("department")).
		Uint("assigned_to", q.QueryParam("assigned_to"))
}
