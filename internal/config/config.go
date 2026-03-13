package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	ErrFmtFailedToLoadConfig = "failed to load config for env %s: %w"
	defaultEnv               = "dev"
	envKey                   = "APP_ENV"
	configDir                = "config"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	Deepgram  DeepgramConfig  `mapstructure:"deepgram"`
	LLM       LLMConfig       `mapstructure:"llm"`
	Salesforce SalesforceConfig `mapstructure:"salesforce"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Tracing   TracingConfig   `mapstructure:"tracing"`
}

type AppConfig struct {
	Env             string `mapstructure:"env"`
	Name            string `mapstructure:"name"`
	Port            string `mapstructure:"port"`
	OpsPort         string `mapstructure:"ops_port"`
	ShutdownDelay   int    `mapstructure:"shutdown_delay"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string              `mapstructure:"host"`
	Port            int                 `mapstructure:"port"`
	User            string              `mapstructure:"user"`
	Password        string              `mapstructure:"password"`
	Name            string              `mapstructure:"name"`
	SSLMode         string              `mapstructure:"ssl_mode"`
	MaxConnections  int32               `mapstructure:"max_connections"`
	MinConnections  int32               `mapstructure:"min_connections"`
	MaxConnLifetime string              `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime string              `mapstructure:"max_conn_idle_time"`
	Retry           DatabaseRetryConfig `mapstructure:"retry"`
}

type DatabaseRetryConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	MaxRetries     int    `mapstructure:"max_retries"`
	InitialBackoff string `mapstructure:"initial_backoff"`
	MaxBackoff     string `mapstructure:"max_backoff"`
}

type RedisConfig struct {
	Addr       string `mapstructure:"addr"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	SessionTTL string `mapstructure:"session_ttl"`
}

func (r *RedisConfig) GetSessionTTL() time.Duration {
	d, err := time.ParseDuration(r.SessionTTL)
	if err != nil {
		return 30 * time.Minute
	}
	return d
}

type KafkaConfig struct {
	Brokers          []string `mapstructure:"brokers"`
	GroupID          string   `mapstructure:"group_id"`
	TopicsAudio      string   `mapstructure:"topics_audio"`
	TopicsTranscript string   `mapstructure:"topics_transcript"`
	TopicsEntities   string   `mapstructure:"topics_entities"`
	TopicsFields     string   `mapstructure:"topics_fields"`
	TopicsEvents     string   `mapstructure:"topics_events"`
}

type DeepgramConfig struct {
	APIKey          string   `mapstructure:"api_key"`
	WSURL           string   `mapstructure:"ws_url"`
	Model           string   `mapstructure:"model"`
	Language        string   `mapstructure:"language"`
	SmartFormat     bool     `mapstructure:"smart_format"`
	Diarize         bool     `mapstructure:"diarize"`
	InterimResults  bool     `mapstructure:"interim_results"`
	CustomVocabulary []string `mapstructure:"custom_vocabulary"`
}

type LLMConfig struct {
	Provider         string `mapstructure:"provider"`
	Model            string `mapstructure:"model"`
	Region           string `mapstructure:"region"`
	MaxTokens        int    `mapstructure:"max_tokens"`
	Temperature      float64 `mapstructure:"temperature"`
	Timeout          string `mapstructure:"timeout"`
	FallbackProvider string `mapstructure:"fallback_provider"`
	FallbackModel    string `mapstructure:"fallback_model"`
}

func (l *LLMConfig) GetTimeout() time.Duration {
	d, err := time.ParseDuration(l.Timeout)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

type SalesforceConfig struct {
	InstanceURL  string `mapstructure:"instance_url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	APIVersion   string `mapstructure:"api_version"`
	ObjectName   string `mapstructure:"object_name"`
	Timeout      string `mapstructure:"timeout"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

func (s *SalesforceConfig) GetTimeout() time.Duration {
	d, err := time.ParseDuration(s.Timeout)
	if err != nil {
		return 10 * time.Second
	}
	return d
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type RateLimitConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	RequestsPerSec float64 `mapstructure:"requests_per_second"`
	BurstSize      int     `mapstructure:"burst_size"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	Endpoint     string  `mapstructure:"endpoint"`
	ServiceName  string  `mapstructure:"service_name"`
	SampleRate   float64 `mapstructure:"sample_rate"`
	Insecure     bool    `mapstructure:"insecure"`
	BatchTimeout string  `mapstructure:"batch_timeout"`
}

var C *Config

func Load() (*Config, error) {
	env := os.Getenv(envKey)
	if env == "" {
		env = defaultEnv
	}
	return LoadForEnv(env)
}

func LoadForEnv(env string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	v.SetConfigName("default")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf(ErrFmtFailedToLoadConfig, env, err)
	}

	v.SetConfigName(env)
	if err := v.MergeInConfig(); err != nil {
		// env-specific config is optional
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf(ErrFmtFailedToLoadConfig, env, err)
	}

	C = &cfg
	return &cfg, nil
}

func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *DatabaseConfig) GetMaxConnLifetime() time.Duration {
	d, err := time.ParseDuration(c.MaxConnLifetime)
	if err != nil {
		return time.Hour
	}
	return d
}

func (c *DatabaseConfig) GetMaxConnIdleTime() time.Duration {
	d, err := time.ParseDuration(c.MaxConnIdleTime)
	if err != nil {
		return 30 * time.Minute
	}
	return d
}
