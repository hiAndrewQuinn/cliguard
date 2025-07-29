package inspector

// InspectedCLI represents the actual CLI structure found by inspection
type InspectedCLI struct {
	Use      string             `json:"use"`
	Short    string             `json:"short"`
	Long     string             `json:"long,omitempty"`
	Flags    []InspectedFlag    `json:"flags,omitempty"`
	Commands []InspectedCommand `json:"commands,omitempty"`
}

// InspectedCommand represents an actual subcommand found by inspection
type InspectedCommand struct {
	Use      string             `json:"use"`
	Short    string             `json:"short"`
	Long     string             `json:"long,omitempty"`
	Flags    []InspectedFlag    `json:"flags,omitempty"`
	Commands []InspectedCommand `json:"commands,omitempty"`
}

// InspectedFlag represents an actual flag found by inspection
type InspectedFlag struct {
	Name       string `json:"name"`
	Shorthand  string `json:"shorthand,omitempty"`
	Usage      string `json:"usage"`
	Type       string `json:"type"`
	Persistent bool   `json:"persistent"`
}
