// Package config provides configuration loading with environment-specific overrides.
//
// Usage:
//
//	Application should have a directory holding default.toml and environment
//	specific files (dev.toml, test.toml, prod.toml).
//	Use NewDefaultConfig().Load("dev", &config, "APP") to load config for dev environment.
package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultConfigType     = "toml"
	DefaultConfigDir      = "./config"
	DefaultConfigFileName = "default"
	WorkDirEnv            = "WORKDIR"
	AppEnvKey             = "APP_ENV"
	DefaultEnv            = "dev"
)

// Options holds configuration loading options
type Options struct {
	configType            string
	configPath            string
	defaultConfigFileName string
}

// Config wraps viper for configuration loading
type Config struct {
	opts  Options
	viper *viper.Viper
}

// NewDefaultOptions returns default configuration options.
// Uses WORKDIR env var if set, otherwise resolves relative to the caller.
func NewDefaultOptions() Options {
	configPath := resolveConfigPath()
	return NewOptions(DefaultConfigType, configPath, DefaultConfigFileName)
}

func resolveConfigPath() string {
	workDir := os.Getenv(WorkDirEnv)
	if workDir != "" {
		return path.Join(workDir, DefaultConfigDir)
	}

	_, thisFile, _, ok := runtime.Caller(2)
	if !ok {
		return DefaultConfigDir
	}

	return path.Join(path.Dir(thisFile), "../../"+DefaultConfigDir)
}

// NewOptions creates new Options with specified values
func NewOptions(configType string, configPath string, defaultConfigFileName string) Options {
	return Options{
		configType:            configType,
		configPath:            configPath,
		defaultConfigFileName: defaultConfigFileName,
	}
}

// NewDefaultConfig creates a new Config with default options
func NewDefaultConfig() *Config {
	return NewConfig(NewDefaultOptions())
}

// NewConfig creates a new Config with specified options
func NewConfig(opts Options) *Config {
	return &Config{
		opts:  opts,
		viper: viper.New(),
	}
}

// Load reads default configuration and environment-specific overrides,
// then unmarshals into the provided config struct.
func (c *Config) Load(env string, config interface{}, prefix string) error {
	if err := c.loadByConfigName(c.opts.defaultConfigFileName, config, prefix); err != nil {
		return err
	}
	return c.loadByConfigName(env, config, prefix)
}

func (c *Config) loadByConfigName(configName string, config interface{}, prefix string) error {
	c.viper.SetEnvPrefix(strings.ToUpper(prefix))
	c.viper.SetConfigName(configName)
	c.viper.SetConfigType(c.opts.configType)
	c.viper.AddConfigPath(c.opts.configPath)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := c.viper.ReadInConfig(); err != nil {
		return err
	}

	configFile := c.viper.ConfigFileUsed()

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	expandedContent := expandEnvVars(content)

	if err := c.viper.ReadConfig(strings.NewReader(string(expandedContent))); err != nil {
		return err
	}

	return c.viper.Unmarshal(config)
}

// GetEnv returns the current environment from APP_ENV or default
func GetEnv() string {
	env := os.Getenv(AppEnvKey)
	if env == "" {
		return DefaultEnv
	}
	return env
}
