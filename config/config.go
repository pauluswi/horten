package config

// Config holds the application configuration.
type Config struct {
	Port string
}

// LoadConfig loads configuration for the application.
func LoadConfig() *Config {
	return &Config{
		Port: "8080",
	}
}
