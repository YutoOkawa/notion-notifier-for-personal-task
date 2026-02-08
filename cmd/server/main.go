package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/application"
	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/config"
	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/infrastructure/discord"
	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/infrastructure/notion"
	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "/etc/config/notion-notifier/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	notionClient := notion.NewClient(cfg.Notion.APIToken, cfg.Notion.DatabaseID)
	discordClient := discord.NewWebhookClient(cfg.Discord.WebhookURL)
	notificationService := application.NewNotificationService(notionClient, discordClient, cfg.Notification.DaysBefore)

	schedule := cfg.Notification.CheckSchedule
	if schedule == "" {
		schedule = "0 12 * * *"
	}

	s := scheduler.New(schedule, notificationService)
	if err := s.Start(); err != nil {
		log.Fatalf("failed to start scheduler: %v", err)
	}

	if os.Getenv("RUN_ON_STARTUP") == "true" {
		if err := s.RunNow(); err != nil {
			log.Printf("initial run error: %v", err)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	s.Stop()
}
