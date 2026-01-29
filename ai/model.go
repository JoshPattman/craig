package ai

import (
	"encoding/json"
	"time"

	"github.com/JoshPattman/react"

	"github.com/JoshPattman/jpf"
	"github.com/invopop/jsonschema"
)

func NewModelBuilder(apiKey string, model string) react.ModelBuilder {
	return &simpleAgentModelBuilder{
		APIKey: apiKey,
		Model:  model,
	}
}

type simpleAgentModelBuilder struct {
	APIKey string
	Model  string
}

// BuildFragmentSelectorModel implements [react.ModelBuilder].
func (m *simpleAgentModelBuilder) BuildFragmentSelectorModel(responseType any) jpf.Model {
	return m.BuildAgentModel(responseType, nil, nil)
}

func (m *simpleAgentModelBuilder) BuildAgentModel(responseType any, onInitFinalStream func(), onDataFinalStream func(string)) jpf.Model {
	var schema map[string]any
	if responseType != nil {
		var err error
		schema, err = getSchema(responseType)
		if err != nil {
			panic(err)
		}
	}
	model := jpf.NewOpenAIModel(
		m.APIKey,
		m.Model,
		jpf.WithJsonSchema{X: schema},
		jpf.WithStreamResponse{OnBegin: onInitFinalStream, OnText: onDataFinalStream},
	)
	model = jpf.NewRetryModel(model, 5, jpf.WithDelay{X: time.Second})
	//model = jpf.NewLoggingModel(model, jpf.NewSlogModelLogger(slog.Info, false))
	return model
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
