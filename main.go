package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pragmader/q3-server-bot/pkg/events"
	"github.com/pragmader/q3-server-bot/pkg/rcon"
	"github.com/pragmader/q3-server-bot/pkg/telegram"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.SetLevel(logrus.DebugLevel)

	token := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	password := os.Getenv("QUAKE3_PASSWORD")
	serverAddr := os.Getenv("QUAKE3_SERVER")

	sender := rcon.NewSender(serverAddr, password)
	em := events.NewManager(sender, 5*time.Second, 5*time.Second)
	bot := telegram.NewServerEventsBot(em, token, chatID, serverAddr, 3*time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		em.Close()
		bot.Close()
		cancel()
		os.Exit(1)
	}()

	go func() {
		err := bot.Start(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	err := em.StartCapturing(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
}
