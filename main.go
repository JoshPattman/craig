package main

import (
	"log/slog"
	"os"
	"strings"
)

func main() {
	logger := slog.Default()
	app, err := NewApp(
		os.Getenv("OPENAI_KEY"),
		os.Getenv("GEMINI_KEY"),
		logger, "/craig-data/agent",
		strings.TrimSpace(strings.ToLower(os.Getenv("CRAIG_INIT"))) == "yes",
	)
	if err != nil {
		panic(err)
	}
	session, err := NewSession(app, os.Getenv("CRAIG_DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}
	err = RunSession(session, logger)
	if err != nil {
		panic(err)
	}
}
