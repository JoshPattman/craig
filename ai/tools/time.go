package tools

import (
	"time"

	"github.com/JoshPattman/react"
)

func NewTimeTool() react.Tool {
	return &timeTool{}
}

type timeTool struct {
}

func (t *timeTool) Call(map[string]any) (string, error) {
	return time.Now().Format(time.ANSIC), nil
}

func (t *timeTool) Name() string {
	return "get_time"
}

func (t *timeTool) Description() []string {
	return []string{
		"Gets the current time",
		"Takes no arguments",
	}
}
