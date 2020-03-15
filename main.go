package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pragmader/q3-server-bot/pkg/parsers"
	"github.com/pragmader/q3-server-bot/pkg/rcon"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logrus.SetLevel(logrus.DebugLevel)

	password := os.Getenv("QUAKE3_PASSWORD")
	serverAddr := os.Getenv("QUAKE3_SERVER")

	command := "players"
	sender := rcon.NewSender(serverAddr, password)
	output, err := sender.Send(ctx, command)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("command output:\n%s\n", output)

	players := parsers.ParsePlayers(output)
	fmt.Printf("parsed output:\n%v\n", players)
}
