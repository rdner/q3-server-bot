package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pragmader/q3-server-bot/pkg/events"
	"github.com/pragmader/q3-server-bot/pkg/rcon"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.SetLevel(logrus.DebugLevel)

	password := os.Getenv("QUAKE3_PASSWORD")
	serverAddr := os.Getenv("QUAKE3_SERVER")

	sender := rcon.NewSender(serverAddr, password)

	em := events.NewManager(sender, 5*time.Second, 5*time.Second)

	events := em.Subscribe()

	go func() {
		err := em.StartCapturing(ctx)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		em.Close()
		os.Exit(1)
	}()

	for e := range events {
		logrus.Infof("%T: %+v", e, e)
	}
	logrus.Info("subscription closed")
}
