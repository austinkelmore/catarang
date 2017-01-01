package template

import (
	"encoding/json"
)

type Job struct {
	Steps     []Step
	LocalPath string
}

type Step struct {
	PluginName string          `json:"plugin_name"`
	PluginData json.RawMessage `json:"plugin"`
}
