package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Notion       NotionConfig       `yaml:"notion"`
	Discord      DiscordConfig      `yaml:"discord"`
	Notification NotificationConfig `yaml:"notification"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type NotionConfig struct {
	APIToken   string `yaml:"api_token"`
	DatabaseID string `yaml:"database_id"`
}

type DiscordConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

type NotificationConfig struct {
	DaysBefore    int    `yaml:"days_before"`
	CheckSchedule string `yaml:"check_schedule"` // cron形式: "0 12 * * *" = 毎日12時
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 環境変数を展開
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Notion.APIToken == "" {
		return fmt.Errorf("notion.api_token is required")
	}
	if c.Notion.DatabaseID == "" {
		return fmt.Errorf("notion.database_id is required")
	}
	if c.Discord.WebhookURL == "" {
		return fmt.Errorf("discord.webhook_url is required")
	}
	return nil
}
