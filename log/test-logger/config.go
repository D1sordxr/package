package test_logger

type Config struct {
	LogLevel         string   `yaml:"log_level"`
	IsFormatted      bool     `yaml:"is_formatted"`
	ContextLogFields []string `yaml:"context_log_fields"`
	CallerSkip       int      `yaml:"caller_skip"`
	BufferSize       int      `yaml:"buffer_size"`
	OverflowStrategy BufferOverflowStrategy
}
