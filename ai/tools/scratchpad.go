package tools

import (
	"errors"
	"os"
	"strings"

	"github.com/JoshPattman/react"
)

func NewScratchPad(filepath string) *ScratchPad {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		os.WriteFile(filepath, []byte(""), os.ModePerm)
	}
	return &ScratchPad{filepath}
}

type ScratchPad struct {
	filepath string
}

func (s *ScratchPad) Content() (string, error) {
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

var ErrOldTextNotFound = errors.New("old text was not found in the scratchpad")
var ErrOldTextAmbiguous = errors.New("old text was ambiguous in the scratchpad")

func (s *ScratchPad) Rewrite(oldText, newText string) error {
	content, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}
	n := strings.Count(string(content), oldText)
	if n == 0 {
		return ErrOldTextNotFound
	}
	if n > 1 {
		return ErrOldTextAmbiguous
	}
	newContent := strings.ReplaceAll(string(content), oldText, newText)
	err = os.WriteFile(s.filepath, []byte(newContent), os.ModePerm)
	return nil
}

func NewReadScratchPadTool(sp *ScratchPad) react.Tool {
	return &readScratchPadTool{sp: sp}
}

type readScratchPadTool struct {
	sp *ScratchPad
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

func NewRewriteScratchPadTool(sp *ScratchPad) react.Tool {
	return &rewriteScratchPadTool{sp: sp}
}

type rewriteScratchPadTool struct {
	sp *ScratchPad
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
