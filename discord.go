package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Runnable interface {
	OnMessageCreate(*discordgo.Session, *discordgo.MessageCreate)
}

func NewSession(app Runnable, botToken string) (*discordgo.Session, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	dg.AddHandler(app.OnMessageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers
	return dg, nil
}

func RunSession(session *discordgo.Session, logger *slog.Logger) error {
	err := session.Open()
	if err != nil {
		logger.Error("Failed to start session", "err", err.Error())
		return err
	}
	defer session.Close()
	logger.Info("Session is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	logger.Info("Session has gracefully quit")
	return nil
}
