package main

import (
	"craig/ai"
	"craig/ai/tools"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

func NewApp(openAIKey string, logger *slog.Logger, scratchpadLocation string) *App {
	app := &App{
		openAIKey:          openAIKey,
		logger:             logger,
		aiLock:             &sync.Mutex{},
		scratchpadLocation: scratchpadLocation,
	}
	app.resetAgent()
	return app
}

type App struct {
	agent              *ai.AgentRuntime
	lastMessage        time.Time
	openAIKey          string
	logger             *slog.Logger
	aiLock             *sync.Mutex
	scratchpadLocation string
}

const internalErrMessage = "There was an error processing this request"

func (app *App) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	sendData, err := app.getMessageSendData(s, m)
	if err != nil {
		app.logger.Error("Failed to get message send data", "err", err.Error())
		s.ChannelMessageSend(m.ChannelID, internalErrMessage)
		return
	}
	app.logger.Info("Message received", "from", sendData.authorName, "location", sendData.LocationString())
	response, err := app.getAgentResponseHelper(m.Content, sendData.authorName, sendData.LocationString())
	if err != nil {
		app.logger.Error("Failed to call agent", "err", err.Error())
		s.ChannelMessageSend(m.ChannelID, internalErrMessage)
		return
	}
	app.logger.Info("Response generated", "len", len(response))
	if len(response) == 0 {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		app.logger.Error("Failed to send response", "err", err.Error())
		s.ChannelMessageSend(m.ChannelID, internalErrMessage)
		return
	}
	app.logger.Info("Replied")
}

func (app *App) getAgentResponseHelper(msg string, author string, location string) (string, error) {
	if time.Since(app.lastMessage) > time.Hour {
		app.resetAgent()
		app.lastMessage = time.Now()
		app.logger.Info("Resetting agent due to long time since last conversation")
	}
	app.aiLock.Lock()
	response, err := app.agent.Send(msg, author, location)
	app.aiLock.Unlock()
	if err != nil {
		return "", err
	}
	return response, nil
}

type messageSendData struct {
	authorName          string
	channelName         string
	guildName           string
	conversationMembers int
}

func (d messageSendData) LocationString() string {
	return fmt.Sprintf("Discord(server='%s', channel='%s', channel_n_members_including_you=%d)", d.guildName, d.channelName, d.conversationMembers)
}

func (app *App) getMessageSendData(s *discordgo.Session, m *discordgo.MessageCreate) (messageSendData, error) {
	name := m.Author.DisplayName()
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		app.logger.Error("Failed to get channel", "err", err.Error())
		s.ChannelMessageSend(m.ChannelID, fmt.Sprint(err))
		return messageSendData{}, err
	}
	guild, err := s.GuildWithCounts(channel.GuildID)
	if err != nil {
		app.logger.Error("Failed to get guild", "err", err.Error())
		s.ChannelMessageSend(m.ChannelID, fmt.Sprint(err))
		return messageSendData{}, err
	}
	return messageSendData{
		authorName:          name,
		channelName:         channel.Name,
		guildName:           guild.Name,
		conversationMembers: guild.ApproximateMemberCount,
	}, nil
}

func (app *App) resetAgent() {
	modelBuilder := ai.NewModelBuilder(app.openAIKey, "gpt-4.1")
	agentBuilder := ai.NewAgentBuilder(modelBuilder, tools.NewScratchPad(app.scratchpadLocation))
	app.agent = agentBuilder.BuildNew()
}
