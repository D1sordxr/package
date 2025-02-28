package logger

type Config struct {
	LogLevel         string
	ContextLogFields []string `mapstructure:"context_log_fields"`
	CallerSkip       int
	BufferSize       int
}
