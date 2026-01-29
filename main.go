package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.Default()
	app := NewApp(os.Getenv("OPENAI_KEY"), logger, "/craig-data/scratchpad.txt")
	session, err := NewSession(app, os.Getenv("CRAIG_DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}
	err = RunSession(session, logger)
	if err != nil {
		panic(err)
	}
}
