package jobdata

import (
	"encoding/json"
)

type JobTemplate struct {
	Steps     []StepTemplate
	LocalPath string
}

type StepTemplate struct {
	PluginName string          `json:"plugin_name"`
	PluginData json.RawMessage `json:"plugin"`
}
