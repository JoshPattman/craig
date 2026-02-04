package ai

import (
	"craig/data"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/JoshPattman/react"

	"github.com/JoshPattman/jpf"
	"github.com/invopop/jsonschema"
)

func NewModelBuilder(agentSetup, filterSetup data.ModelSetup, openAIKey, geminiKey string) react.ModelBuilder {
	return &simpleAgentModelBuilder{
		agentSetup, filterSetup, openAIKey, geminiKey,
	}
}

type simpleAgentModelBuilder struct {
	agentSetup  data.ModelSetup
	filterSetup data.ModelSetup
	openAIKey   string
	geminiKey   string
}

func (m *simpleAgentModelBuilder) BuildFragmentSelectorModel(responseType any) jpf.Model {
	model, err := buildModel(m.filterSetup, m.openAIKey, m.geminiKey, responseType, nil, nil)
	if err != nil {
		panic(err)
	}
	return model
}

func (m *simpleAgentModelBuilder) BuildAgentModel(responseType any, onInitFinalStream func(), onDataFinalStream func(string)) jpf.Model {
	model, err := buildModel(m.agentSetup, m.openAIKey, m.geminiKey, responseType, onInitFinalStream, onDataFinalStream)
	if err != nil {
		panic(err)
	}
	return model
}

func buildModel(setup data.ModelSetup, openAIKey, geminiKey string, responseType any, onInitFinalStream func(), onDataFinalStream func(string)) (jpf.Model, error) {
	var model jpf.Model
	switch setup.Provider {
	case "openai":
		args := []jpf.OpenAIModelOpt{
			jpf.WithURL{X: setup.URL},
			jpf.WithStreamResponse{OnBegin: onInitFinalStream, OnText: onDataFinalStream},
		}
		if setup.Headers != nil {
			for k, v := range setup.Headers {
				args = append(args, jpf.WithHTTPHeader{K: k, V: v})
			}
		}
		if responseType != nil {
			schema, err := getSchema(responseType)
			if err != nil {
				panic(err)
			}
			args = append(args, jpf.WithJsonSchema{X: schema})
		}
		if setup.Temperature != nil {
			args = append(args, jpf.WithTemperature{X: *setup.Temperature})
		}
		if setup.ReasoningEffort != nil {
			var re jpf.ReasoningEffort
			switch *setup.ReasoningEffort {
			case "low":
				re = jpf.LowReasoning
			case "medium":
				re = jpf.MediumReasoning
			case "high":
				re = jpf.HighReasoning
			default:
				return nil, fmt.Errorf("unrecognised reasoning effort '%s'", *setup.ReasoningEffort)
			}
			args = append(args, jpf.WithReasoningEffort{X: re})
		}
		model = jpf.NewOpenAIModel(openAIKey, setup.Name, args...)

	case "gemini":
		args := []jpf.GeminiModelOpt{
			jpf.WithURL{X: setup.URL},
			jpf.WithStreamResponse{OnBegin: onInitFinalStream, OnText: onDataFinalStream},
		}
		if setup.Headers != nil {
			for k, v := range setup.Headers {
				args = append(args, jpf.WithHTTPHeader{K: k, V: v})
			}
		}
		if setup.Temperature != nil {
			args = append(args, jpf.WithTemperature{X: *setup.Temperature})
		}
		model = jpf.NewGeminiModel(geminiKey, setup.Name, args...)
	default:
		return nil, fmt.Errorf("unrecognised provider '%s'", setup.Provider)
	}
	model = jpf.NewLoggingModel(model, jpf.NewSlogModelLogger(slog.Info, false))
	if setup.Retries > 0 {
		model = jpf.NewRetryModel(model, setup.Retries)
	}
	return model, nil
}

func getSchema(obj any) (map[string]any, error) {
	r := &jsonschema.Reflector{
		BaseSchemaID:   "Anonymous",
		Anonymous:      true,
		DoNotReference: true,
	}
	s := r.Reflect(obj)
	schemaBs, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	schema := make(map[string]any)
	err = json.Unmarshal(schemaBs, &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
