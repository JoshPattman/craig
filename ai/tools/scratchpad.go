package tools

import (
	"craig/data"
	"errors"

	"github.com/JoshPattman/react"
)

func NewReadScratchPadTool(sp data.ScratchPad) react.Tool {
	return &readScratchPadTool{sp: sp}
}

type readScratchPadTool struct {
	sp data.ScratchPad
}

func (t *readScratchPadTool) Call(map[string]any) (string, error) {
	return t.sp.Content()
}

func (t *readScratchPadTool) Name() string {
	return "read_scratchpad"
}

func (t *readScratchPadTool) Description() []string {
	return []string{
		"Reads the entire scratchpad",
		"Takes no arguments",
	}
}

func NewRewriteScratchPadTool(sp data.ScratchPad) react.Tool {
	return &rewriteScratchPadTool{sp: sp}
}

type rewriteScratchPadTool struct {
	sp data.ScratchPad
}

func (t *rewriteScratchPadTool) Call(args map[string]any) (string, error) {
	oldText, ok := args["old_text"].(string)
	if !ok {
		return "", errors.New("missing or invalid 'old_text'")
	}

	newText, ok := args["new_text"].(string)
	if !ok {
		return "", errors.New("missing or invalid 'new_text'")
	}

	if err := t.sp.Rewrite(oldText, newText); err != nil {
		return "", err
	}

	return "scratchpad updated", nil
}

func (t *rewriteScratchPadTool) Name() string {
	return "rewrite_scratchpad"
}

func (t *rewriteScratchPadTool) Description() []string {
	return []string{
		"Rewrites part of the scratchpad",
		"Arguments:",
		"- old_text: text to replace (may be empty if scratchpad is empty)",
		"- new_text: replacement text",
	}
}
