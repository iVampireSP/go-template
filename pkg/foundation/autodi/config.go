package main

// Config holds autodi configuration, populated from conventions and generate.go annotations.
type Config struct {
	Module   string
	Scan     []string
	Exclude  []string
	Output   string
	Bindings map[string][]string    // concrete type → interface list (from //autodi:bind)
	Groups   map[string]GroupConfig // from //autodi:group

	// From //autodi:app annotation
	AppName  string
	AppShort string
	AppLong  string
}

// GroupConfig defines a collection of providers implementing an interface.
type GroupConfig struct {
	Interface string
	Paths     []string
}
