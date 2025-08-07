package contract

// Contract represents the complete CLI contract specification.
// It defines the expected structure of a CLI application including
// its root command, flags, and all subcommands.
//
// Example YAML:
//
//	use: myapp
//	short: My application
//	long: |
//	  My application is a tool that does amazing things.
//	  It supports multiple subcommands and flags.
//	flags:
//	  - name: config
//	    type: string
//	    usage: Config file path
//	commands:
//	  - use: serve
//	    short: Start the server
type Contract struct {
	// Use is the command name as it appears when invoked (required).
	// For the root command, this is the application name.
	// Example: "git" for the git CLI
	Use string `yaml:"use"`
	
	// Short is a brief one-line description shown in help listings (required).
	// Should be concise and start with a capital letter.
	// Example: "Fast, scalable, distributed revision control system"
	Short string `yaml:"short"`
	
	// Long is a detailed description shown in the help command (optional).
	// Can be multiple paragraphs and include usage examples.
	Long string `yaml:"long,omitempty"`
	
	// Flags defines the command-line flags available on this command (optional).
	// These are flags specific to this command, not inherited by subcommands
	// unless marked as Persistent.
	Flags []Flag `yaml:"flags,omitempty"`
	
	// Commands lists all subcommands available under this command (optional).
	// Each subcommand can have its own flags and nested subcommands.
	Commands []Command `yaml:"commands,omitempty"`
}

// Command represents a subcommand in the contract.
// Commands can be nested to create complex CLI structures with
// multiple levels of subcommands.
//
// Example YAML:
//
//	use: serve
//	short: Start the server
//	long: |
//	  The serve command starts the application server.
//	  It listens on the specified port and serves requests.
//	flags:
//	  - name: port
//	    type: int
//	    usage: Port to listen on
//	commands:
//	  - use: http
//	    short: Start HTTP server
type Command struct {
	// Use is the command name and usage pattern (required).
	// Can include arguments: "serve <port>" or just the name: "serve"
	// Example: "clone [flags] <repository> [<directory>]"
	Use string `yaml:"use"`
	
	// Short is a brief one-line description for command listings (required).
	// Should be concise and start with a capital letter.
	// Example: "Clone a repository into a new directory"
	Short string `yaml:"short"`
	
	// Long is a detailed description shown in the help command (optional).
	// Can include multiple paragraphs, usage examples, and notes.
	Long string `yaml:"long,omitempty"`
	
	// Flags defines command-specific flags (optional).
	// These flags are only available when this command is invoked.
	Flags []Flag `yaml:"flags,omitempty"`
	
	// Commands lists nested subcommands under this command (optional).
	// Allows building complex command hierarchies.
	// Example: "git remote add" where "add" is nested under "remote"
	Commands []Command `yaml:"commands,omitempty"`
}

// Flag represents a command flag in the contract.
// Flags can be either local to a command or persistent (inherited by subcommands).
//
// Example YAML:
//
//	name: config
//	shorthand: c
//	usage: Config file path
//	type: string
//	persistent: false
//
// Supported types include all pflag types:
//   - Basic: bool, string, int, int8, int16, int32, int64
//   - Unsigned: uint, uint8, uint16, uint32, uint64
//   - Float: float32, float64
//   - Duration: duration (time.Duration)
//   - Slices: stringSlice, intSlice, boolSlice
//   - Special: count (incremental counter)
type Flag struct {
	// Name is the long form of the flag (required).
	// Used with double dash: --name
	// Example: "verbose" for --verbose
	Name string `yaml:"name"`
	
	// Shorthand is the single-letter abbreviation (optional).
	// Used with single dash: -s
	// Example: "v" for -v
	Shorthand string `yaml:"shorthand,omitempty"`
	
	// Usage is the help text shown for this flag (required).
	// Should be concise and describe what the flag does.
	// Example: "Enable verbose output"
	Usage string `yaml:"usage"`
	
	// Type specifies the flag's data type (required).
	// Must be a valid pflag type name.
	// Common types: string, bool, int, float64, duration, stringSlice
	Type string `yaml:"type"`
	
	// Persistent indicates if the flag is inherited by subcommands (optional).
	// When true, this flag is available to all nested subcommands.
	// Default: false (flag is local to the command)
	Persistent bool `yaml:"persistent,omitempty"`
}
