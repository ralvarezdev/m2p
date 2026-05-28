package cli

// Viper config key names
const (
	KeyFormat     = "format"
	KeyPaper      = "paper"
	KeyEngine     = "engine"
	KeyNoFooter   = "no-footer"
	KeyOpen       = "open"
	KeyPageBreak  = "page-break"
	KeyQuiet      = "quiet"
)

// Environment variable prefix
const EnvPrefix = "M2P"

// Default flag values
const (
	DefaultFormat    = "pdf"
	DefaultPaper     = "a4"
	DefaultEngine    = "auto"
	DefaultPageBreak = "none"
)

// Config file defaults
const (
	ConfigFileName = "config"
	ConfigFileType = "toml"
)

// Config directory names
const ConfigDirName = "m2p"

// Environment variables
const (
	EnvAPPDATA        = "APPDATA"
	EnvXDGConfigHome  = "XDG_CONFIG_HOME"
)

// Config file paths
const ConfigPathDefault = ".config"

// OS identifiers
const (
	OSWindows = "windows"
	OSDarwin  = "darwin"
)

// File open commands
const (
	OpenCmdWindows = "rundll32"
	OpenArgWindows = "url.dll,FileProtocolHandler"
	OpenCmdDarwin  = "open"
	OpenCmdLinux   = "xdg-open"
)
