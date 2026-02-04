package main

import (
	"craig/ai"
	"craig/data"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "embed"

	"github.com/JoshPattman/react"
	"github.com/bwmarrin/discordgo"
)

//go:embed defaults
var defaultSetup embed.FS

func NewApp(openAIKey, geminiKey string, logger *slog.Logger, dataLocation string, initialise bool) (*App, error) {
	dd := data.NewDirectoryData(dataLocation)

	if initialise {
		err := dd.Init(defaultSetup, "defaults")
		if err != nil {
			return nil, err
		}
		logger.Info("Extracted default data")
	}

	agentSetup, err := dd.AgentModel()
	if err != nil {
		return nil, err
	}

	filterSetup, err := dd.FilterModel()
	if err != nil {
		return nil, err
	}

	modelBuilder := ai.NewModelBuilder(agentSetup, filterSetup, openAIKey, geminiKey)
	agentBuilder := ai.NewAgentBuilder(
		modelBuilder,
		dd.GetScratchPad(),
		dd.GetSkillset(),
	)

	app := &App{
		logger:       logger,
		aiLock:       &sync.Mutex{},
		agentBuilder: agentBuilder,
	}
	err = app.resetAgent()
	if err != nil {
		return nil, err
	}
	return app, nil
}

type App struct {
	agent        *ai.AgentRuntime
	lastMessage  time.Time
	logger       *slog.Logger
	aiLock       *sync.Mutex
	agentBuilder *ai.AgentBuilder
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
		err := app.resetAgent()
		if err != nil {
			return "", err
		}
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

func (app *App) resetAgent() (err error) {
	app.agent, err = app.agentBuilder.BuildNew()
	return err
}

func loadSkills(skillLocation string) ([]react.Skill, error) {
	var skills []react.Skill
	f, err := os.Open(skillLocation)
	if errors.Is(err, os.ErrNotExist) {
		skills = []react.Skill{
			{
				Key:     "scratch_pad_info",
				Content: "The scratch pad is a block of text, private to you, which you can read and rewrite at will. You should **always** read the scratch pad at the start of a conversation, but you may also read it again through the conversation if you feel you need to (you usually only need to read it once though). The scratch pad will persist across conversations, so you should use it to store information that you deem useful in future conversations. Don't use it to store silly trivial things, but instead things that you think would be useful for you or the user to remember (for example, does a user want you to call them a different name, what are their preferences). While writing to the scratchpad, remember that you may talk to multiple users, so you may at some point need to organise findings by user. There is no format you need to abide by for the scratchpad, use whatever format you like.",
			},
			{
				Key:     "not_replying",
				Content: "Sometimes, you can choose to not reply. To do this, simply make your final message an empty string. You will need to use your intelligence to figure out when not to apply. For example, if there is only one person in the channel with you, you probably should reply unless its obvious otherwise. However, if there are multiple people chatting, you dont need to reply unless explicitly talked to (and are in a current active conversation).",
			},
			{
				Key:     "writing_poems",
				When:    "The user asks the agent to write a poem",
				Content: "When asked to write a poem, the agent will write a haiku by default as that is short and sweet.",
			},
		}
		f, err := os.Create(skillLocation)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "    ")
		err = enc.Encode(skills)
		if err != nil {
			return nil, err
		}
		return skills, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&skills)
	if err != nil {
		return nil, err
	}
	return skills, nil
}
