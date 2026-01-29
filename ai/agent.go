package ai

import (
	"fmt"

	"craig/ai/tools"

	"github.com/JoshPattman/react"
)

func NewAgentBuilder(modelBuilder react.ModelBuilder, pad *tools.ScratchPad) *AgentBuilder {
	return &AgentBuilder{
		modelBuilder: modelBuilder,
		pad:          pad,
	}
}

type AgentBuilder struct {
	modelBuilder react.ModelBuilder
	pad          *tools.ScratchPad
}

func getFragments() []react.PromptFragment {
	return []react.PromptFragment{
		{
			Key:     "scratch_pad_info",
			Content: "The scratch pad is a block of text, private to you, which you can read and rewrite at will. You should **always** read the scratch pad at the start of a conversation, but you may also read it again through the conversation if you feel you need to (you usually only need to read it once though). The scratch pad will persist across conversations, so you should use it to store information that you deem useful in future conversations. Don't use it to store silly trivial things, but instead things that you think would be useful for you or the user to remember (for example, does a user want you to call them a different name, what are their preferences). While writing to the scratchpad, remember that you may talk to multiple users, so you may at some point need to organise findings by user. There is no format you need to abide by for the scratchpad, use whatever format you like.",
		},
		{
			Key:     "not_replying",
			Content: "Sometimes, you can choose to not reply. To do this, simply make your final message an empty string. You will need to use your intelligence to figure out when not to apply. For example, if there is only one person in the channel with you, you probably should reply unless its obvious otherwise. However, if there are multiple people chatting, you dont need to reply unless explicitly talked to (and are in a current active conversation).",
		},
	}
}

func (ab *AgentBuilder) BuildNew() *AgentRuntime {
	agent := react.NewCraig(
		ab.modelBuilder,
		react.WithCraigTools(tools.NewTimeTool(), tools.NewReadScratchPadTool(ab.pad), tools.NewRewriteScratchPadTool(ab.pad)),
		react.WithCraigFragments(getFragments()...),
		react.WithCraigPersonality("You are Craig (Combined ReAct Intelligent aGent). You are an AI assistant, althout you act like a human. You are witty but not to an annoying degree, and a bit sarcastic. When talking to the user, you will be messaging over dicord, so dont use capitals or punctuation much (talk like a normal discord user would), and feel free to use slang if you need. Usually, your responses should be quite short (for normal conversation, maybe one sentence)."),
	)
	return &AgentRuntime{
		agent: agent,
	}
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
