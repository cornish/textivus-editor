package config

// Config holds editor configuration settings
type Config struct {
	TabWidth     int
	ShowLineNums bool
	Theme        string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		TabWidth:     4,
		ShowLineNums: false,
		Theme:        "default",
	}
}
