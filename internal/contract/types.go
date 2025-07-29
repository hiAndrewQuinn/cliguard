package contract

// Contract represents the complete CLI contract specification
type Contract struct {
	Use      string    `yaml:"use"`
	Short    string    `yaml:"short"`
	Long     string    `yaml:"long,omitempty"`
	Flags    []Flag    `yaml:"flags,omitempty"`
	Commands []Command `yaml:"commands,omitempty"`
}

// Command represents a subcommand in the contract
type Command struct {
	Use      string    `yaml:"use"`
	Short    string    `yaml:"short"`
	Long     string    `yaml:"long,omitempty"`
	Flags    []Flag    `yaml:"flags,omitempty"`
	Commands []Command `yaml:"commands,omitempty"`
}

// Flag represents a command flag in the contract
type Flag struct {
	Name       string `yaml:"name"`
	Shorthand  string `yaml:"shorthand,omitempty"`
	Usage      string `yaml:"usage"`
	Type       string `yaml:"type"` // string, bool, int, int64, float64, etc.
	Persistent bool   `yaml:"persistent,omitempty"`
}
