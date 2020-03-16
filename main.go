package main

import (
	"context"
	"flag"
	"fmt"
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

	var (
		isDebug bool

		telegramToken  string
		telegramChatID string
		q3ServerAddr   string
		q3Password     string
	)

	flag.BoolVar(&isDebug, "debug", false, "Use this flag if you'd like to see detailed logging output")

	flag.StringVar(&telegramToken, "telegram-token", "", "Bot token that you get when you create a Telegram bot. You can use 'TELEGRAM_TOKEN' environment variable instead.")
	flag.StringVar(&telegramChatID, "telegram-chat-id", "", "Unique identifier for the target chat or username of the target channel (in the format @channelusername). You can use 'TELEGRAM_CHAT_ID' environment variable instead.")

	flag.StringVar(&q3ServerAddr, "q3-server-addr", "", "Address of the quake server including port. You can use 'Q3_SERVER_ADDR' environment variable instead")
	flag.StringVar(&q3Password, "q3-password", "", "The rcon password of the server. You can use 'Q3_PASSWORD' environment variable instead")

	flag.Parse()

	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if telegramToken == "" {
		telegramToken = os.Getenv("TELEGRAM_TOKEN")
	}
	if telegramToken == "" {
		fmt.Println("`telegram-token` cannot be empty")
		flag.Usage()
		os.Exit(1)
	}

	if telegramChatID == "" {
		telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
	}
	if telegramChatID == "" {
		fmt.Println("`telegram-chat-id` cannot be empty")
		flag.Usage()
		os.Exit(1)
	}

	if q3ServerAddr == "" {
		q3ServerAddr = os.Getenv("Q3_SERVER_ADDR")
	}
	if q3ServerAddr == "" {
		fmt.Println("`q3-server-addr` cannot be empty")
		flag.Usage()
		os.Exit(1)
	}

	if q3Password == "" {
		q3Password = os.Getenv("Q3_PASSWORD")
	}
	if q3Password == "" {
		fmt.Println("`q3-password` cannot be empty")
		flag.Usage()
		os.Exit(1)
	}

	logrus.Info("initializing...")
	sender := rcon.NewSender(q3ServerAddr, q3Password)
	em := events.NewManager(sender, 5*time.Second, 5*time.Second)
	bot := telegram.NewServerEventsBot(
		em,
		telegramToken,
		telegramChatID,
		q3ServerAddr,
		3*time.Second,
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logrus.Info("starting event capturing and sending messages...")
	go func() {
		err := bot.Start(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	go func() {
		err := em.StartCapturing(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	<-c
	logrus.Info("shutting down events...")
	em.Close()
	logrus.Info("shutting down the bot...")
	bot.Close()
	cancel()
	logrus.Info("all stopped")
	os.Exit(1)
}
