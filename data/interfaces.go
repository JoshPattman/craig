package data

import (
	"errors"

	"github.com/JoshPattman/react"
)

var ErrOldTextNotFound = errors.New("old text was not found in the scratchpad")
var ErrOldTextAmbiguous = errors.New("old text was ambiguous in the scratchpad")

type ScratchPad interface {
	Content() (string, error)
	Rewrite(oldText, newText string) error
}

type Skillset interface {
	List() ([]react.Skill, error)
}

type ModelSetup struct {
	Name            string            `json:"name"`
	URL             string            `json:"url"`
	Provider        string            `json:"provider"`
	Retries         int               `json:"retries"`
	Headers         map[string]string `json:"headers"`
	Temperature     *float64          `json:"temperature"`
	ReasoningEffort *string           `json:"reasoning_effort"`
}

type Models interface {
	AgentModel() (ModelSetup, error)
	FilterModel() (ModelSetup, error)
}

type Personality interface {
	Personality() (string, error)
}

type Tools interface {
	EnabledTools() ([]react.Tool, error)
}
