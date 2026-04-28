package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cicbyte/bark-cli/internal/models"
	"go.yaml.in/yaml/v3"
)

var ConfigInstance = Config{}

type Config struct {
	HomeDir      string
	AppSeriesDir string
	AppDir       string
	ConfigDir    string
	ConfigPath   string
	LogDir       string
	LogPath      string
}

func (c *Config) GetHomeDir() string {
	if c.HomeDir != "" {
		return c.HomeDir
	}
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}
	c.HomeDir = usr.HomeDir
	return c.HomeDir
}

func (c *Config) GetAppSeriesDir() string {
	if c.AppSeriesDir != "" {
		return c.AppSeriesDir
	}
	c.AppSeriesDir = c.GetHomeDir() + "/.cicbyte"
	return c.AppSeriesDir
}

func (c *Config) GetAppDir() string {
	if c.AppDir != "" {
		return c.AppDir
	}
	c.AppDir = c.GetAppSeriesDir() + "/bark-cli"
	return c.AppDir
}

func (c *Config) GetConfigDir() string {
	if c.ConfigDir != "" {
		return c.ConfigDir
	}
	c.ConfigDir = c.GetAppDir() + "/config"
	return c.ConfigDir
}

func (c *Config) GetConfigPath() string {
	if c.ConfigPath != "" {
		return c.ConfigPath
	}
	c.ConfigPath = c.GetConfigDir() + "/config.yaml"
	return c.ConfigPath
}

func (c *Config) GetLogDir() string {
	if c.LogDir == "" {
		c.LogDir = filepath.Join(c.GetAppDir(), "logs")
	}
	return c.LogDir
}

func (c *Config) GetLogPath() string {
	if c.LogPath == "" {
		now := time.Now().Format("20060102")
		c.LogPath = filepath.Join(c.GetLogDir(), fmt.Sprintf("bark-cli_log_%s.log", now))
	}
	return c.LogPath
}

func (c *Config) LoadConfig() *models.AppConfig {
	configPath := c.GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := GetDefaultConfig()
		data, _ := yaml.Marshal(cfg)
		_ = os.WriteFile(configPath, data, 0644)
		return cfg
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return GetDefaultConfig()
	}

	var raw map[string]any
	if yaml.Unmarshal(data, &raw) == nil {
		if server, ok := raw["server"].(map[string]any); ok {
			if _, hasURL := server["url"]; hasURL {
				return migrateOldConfig(data)
			}
		}
	}

	var config models.AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return GetDefaultConfig()
	}
	return &config
}

func (c *Config) SaveConfig(config *models.AppConfig) {
	configPath := c.GetConfigPath()
	data, err := yaml.Marshal(config)
	if err != nil {
		return
	}
	os.WriteFile(configPath, data, 0644)
}

func GetDefaultConfig() *models.AppConfig {
	cfg := &models.AppConfig{}
	cfg.Version = "1.0"
	cfg.Servers = &models.Servers{
		Instances: make(map[string]*models.ServerInstance),
	}
	cfg.Output.Format = "table"
	cfg.Log.Level = "info"
	cfg.Log.MaxSize = 10
	cfg.Log.MaxBackups = 30
	cfg.Log.MaxAge = 30
	cfg.Log.Compress = true
	return cfg
}

type oldConfig struct {
	Version string `yaml:"version"`
	Server  struct {
		URL         string `yaml:"url"`
		Key         string `yaml:"key"`
		DeviceToken string `yaml:"device_token"`
		Token       string `yaml:"token"`
		Timeout     int    `yaml:"timeout"`
	} `yaml:"server"`
	AI struct {
		Provider    string  `yaml:"provider"`
		BaseURL     string  `yaml:"base_url"`
		ApiKey      string  `yaml:"api_key"`
		Model       string  `yaml:"model"`
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     int     `yaml:"timeout"`
	} `yaml:"ai"`
	Output struct {
		Format string `yaml:"format"`
	} `yaml:"output"`
	Log struct {
		Level      string `yaml:"level"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

func migrateOldConfig(data []byte) *models.AppConfig {
	var old oldConfig
	if err := yaml.Unmarshal(data, &old); err != nil {
		return GetDefaultConfig()
	}

	cfg := &models.AppConfig{}
	cfg.Version = old.Version
	cfg.AI = old.AI
	cfg.Output = old.Output
	cfg.Log = old.Log

	if old.Server.URL != "" {
		name := "default"
		cfg.Servers = &models.Servers{
			Default: name,
			Instances: map[string]*models.ServerInstance{
				name: {
					URL:         old.Server.URL,
					DeviceToken: old.Server.DeviceToken,
					DeviceKey:   old.Server.Key,
					Token:       old.Server.Token,
					Timeout:     old.Server.Timeout,
				},
			},
		}
		if cfg.Servers.Instances[name].Timeout <= 0 {
			cfg.Servers.Instances[name].Timeout = 10
		}
	} else {
		cfg.Servers = &models.Servers{Instances: make(map[string]*models.ServerInstance)}
	}

	return cfg
}
