package config

type Config struct {
	LogLevel         string
	ContextLogFields []string `mapstructure:"context_log_fields"`
	CallerSkip       int
}
