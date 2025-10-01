package logger

type LogConfig struct {
	Level  string `json:"level" yaml:"level" env:"LOG_LEVEL"`
	Format string `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Output string `json:"output" yaml:"output" env:"LOG_OUTPUT"`
	Caller bool   `json:"caller" yaml:"caller" env:"LOG_CALLER"`
}

func DevelopmentConfig() *LogConfig {
	return &LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
		Caller: true,
	}
}

func ProductionConfig() *LogConfig {
	return &LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}

func (c *LogConfig) Validate() {
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Level] {
		c.Level = "info"
	}

	if c.Format != "console" && c.Format != "json" {
		c.Format = "json"
	}

	if c.Output != "stdout" && c.Output != "stderr" && c.Output != "file" {
		c.Output = "stdout"
	}
}

func (c *LogConfig) IsDevelopment() bool {
	return c.Format == "console" && (c.Level == "debug" || c.Level == "trace")
}

func (c *LogConfig) IsProduction() bool {
	return c.Format == "json" && c.Level != "debug" && c.Level != "trace"
}
