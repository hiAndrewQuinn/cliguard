package inspector

// InspectedCLI represents the actual CLI structure found by inspection.
// This is the result of analyzing a cobra-based CLI application to extract
// its complete command tree, flags, and metadata.
//
// Example JSON representation:
//
//	{
//	  "use": "myapp",
//	  "short": "My application",
//	  "long": "A longer description...",
//	  "flags": [
//	    {
//	      "name": "config",
//	      "shorthand": "c",
//	      "usage": "Config file",
//	      "type": "string",
//	      "persistent": true
//	    }
//	  ],
//	  "commands": [
//	    {
//	      "use": "serve",
//	      "short": "Start server"
//	    }
//	  ]
//	}
type InspectedCLI struct {
	// Use is the command name as defined in cobra.Command.Use
	Use string `json:"use"`

	// Short is the short description from cobra.Command.Short
	Short string `json:"short"`

	// Long is the long description from cobra.Command.Long
	Long string `json:"long,omitempty"`

	// Flags contains all flags defined on the root command
	Flags []InspectedFlag `json:"flags,omitempty"`
	
	// Aliases contains alternative names for this CLI (omitempty)
	Aliases []string `json:"aliases,omitempty"`
	
	// Example contains usage examples for this CLI (omitempty)
	Example string `json:"example,omitempty"`
	
	// Commands contains all direct subcommands
	Commands []InspectedCommand `json:"commands,omitempty"`
}

// InspectedCommand represents an actual subcommand found by inspection.
// Commands can be nested to any depth, forming a tree structure.
type InspectedCommand struct {
	// Use is the command name and usage string
	Use string `json:"use"`

	// Short is the short description for command listings
	Short string `json:"short"`

	// Long is the detailed help text
	Long string `json:"long,omitempty"`

	// Flags contains command-specific flags (not inherited)
	Flags []InspectedFlag `json:"flags,omitempty"`
	
	// Aliases contains alternative names for this command (omitempty)
	Aliases []string `json:"aliases,omitempty"`
	
	// Example contains usage examples for this command (omitempty)
	Example string `json:"example,omitempty"`
	
	// Commands contains nested subcommands
	Commands []InspectedCommand `json:"commands,omitempty"`
}

// InspectedFlag represents an actual flag found by inspection.
// Flags are extracted from cobra/pflag flag sets with their complete metadata.
type InspectedFlag struct {
	// Name is the long flag name (e.g., "verbose" for --verbose)
	Name string `json:"name"`

	// Shorthand is the single-letter abbreviation (e.g., "v" for -v)
	Shorthand string `json:"shorthand,omitempty"`

	// Usage is the help text for the flag
	Usage string `json:"usage"`

	// Type is the flag's data type (string, bool, int, etc.)
	// Uses simplified names for common pflag types
	Type string `json:"type"`

	// Persistent indicates if the flag is inherited by subcommands
	Persistent bool `json:"persistent"`
}
