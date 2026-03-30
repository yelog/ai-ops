package config

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	AI       AIConfig       `yaml:"ai"`
	Monitor  MonitorConfig  `yaml:"monitor"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type AuthConfig struct {
	JWTSecret      string `yaml:"jwt_secret"`
	JWTExpiryHours int    `yaml:"jwt_expiry_hours"`
}

type AIConfig struct {
	Provider       string `yaml:"provider"`
	APIKey         string `yaml:"api_key"`
	Model          string `yaml:"model"`
	EmbeddingModel string `yaml:"embedding_model"`
}

type MonitorConfig struct {
	PrometheusRetentionDays int `yaml:"prometheus_retention_days"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}
