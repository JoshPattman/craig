package ai

import (
	"fmt"

	"craig/ai/tools"
	"craig/data"

	"github.com/JoshPattman/react"
)

func NewAgentBuilder(modelBuilder react.ModelBuilder, pad data.ScratchPad, skillset data.Skillset, personality data.Personality, tools data.Tools) *AgentBuilder {
	return &AgentBuilder{
		modelBuilder: modelBuilder,
		pad:          pad,
		skillset:     skillset,
		personality:  personality,
		tools:        tools,
	}
}

type AgentBuilder struct {
	modelBuilder react.ModelBuilder
	pad          data.ScratchPad
	skillset     data.Skillset
	personality  data.Personality
	tools        data.Tools
}

func (ab *AgentBuilder) BuildNew() (*AgentRuntime, error) {
	skills, err := ab.skillset.List()
	if err != nil {
		return nil, err
	}
	personality, err := ab.personality.Personality()
	if err != nil {
		return nil, err
	}

	confTools, err := ab.tools.EnabledTools()
	if err != nil {
		return nil, err
	}

	agent := react.NewCraig(
		ab.modelBuilder,
		react.WithCraigTools(tools.NewTimeTool(), tools.NewReadScratchPadTool(ab.pad), tools.NewRewriteScratchPadTool(ab.pad)),
		react.WithCraigTools(confTools...),
		react.WithCraigSkills(skills...),
		react.WithCraigPersonality(personality),
	)
	return &AgentRuntime{
		agent: agent,
	}, nil
}

type AgentRuntime struct {
	lastUserName string
	lastLocation string
	agent        react.Agent
	hasInit      bool
}

func (r *AgentRuntime) Send(msg string, userName string, location string) (string, error) {
	notifications := []react.NotificationMessage{}
	if userName != r.lastUserName {
		notifications = append(notifications, react.NotificationMessage{
			Kind:    "switch_user",
			Content: fmt.Sprintf("The user that is talking to you has changed. The user that is now talking to you is called %s", userName),
		})
		r.lastUserName = userName
	}
	if location != r.lastLocation {
		notifications = append(notifications, react.NotificationMessage{
			Kind:    "switch_location",
			Content: fmt.Sprintf("The location you are about to reply in (and recieve messages in) has changed. The location you are in is now %s", location),
		})
		r.lastLocation = location
	}
	if !r.hasInit {
		r.hasInit = true
		notifications = append(notifications, react.NotificationMessage{
			Kind:    "reminder",
			Content: "Remember to check your scratchpad immediately before anything else (only required on this first message)",
		})
	}
	return r.agent.Send(msg, react.WithNotifications(notifications...))
}
