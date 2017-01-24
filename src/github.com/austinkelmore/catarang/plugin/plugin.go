package plugin

import (
	"encoding/json"
	"reflect"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
)

// JobStep is the interface that defines the least amount needed to
// get a plugin working as part of a job
type JobStepper interface {
	Run(logger *cmd.Log) error
	GetName() string
}

// SCM is an interface for the source control modules
type SCM interface {
	FirstTimeSetup(logger *cmd.Log) error
	Poll(logger *cmd.Log) (bool, error)
	UpdateExisting(logger *cmd.Log) error
	SetOrigin(origin string) error
	JobStepper
}

// MarshalJSON will take the JobStep and turn it into the custom JSON
func (s *JobStep) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(s.JobStepper)
	if err != nil {
		return nil, errors.Wrapf(err, "can't marshal JobStep with name %v", s.GetName())
	}
	return json.Marshal(struct {
		PluginName string          `json:"plugin_name"`
		PluginData json.RawMessage `json:"plugin_data"`
	}{
		PluginName: s.GetName(),
		PluginData: data,
	})
}

// JobStep holds the interface for JobStepper and handles the custom serialization needed to
// do run time changes to which interface is used for the job step based off of
// "plugin_name"
type JobStep struct {
	JobStepper
}

func (s *JobStep) UnmarshalJSON(b []byte) error {
	parsed, err := gabs.ParseJSON(b)
	if err != nil {
		return errors.Wrap(err, "error parsing JSON while unmarshaling it")
	}

	pluginNameJSON := parsed.Search("plugin_name")
	if pluginNameJSON.Data() == nil {
		return errors.New("couldn't find \"plugin_name\" in Step")
	}

	name, ok := pluginNameJSON.Data().(string)
	if !ok {
		return errors.New("\"plugin_name\" does not reference a string")
	}

	plug, ok := GetAvailable()[name]
	if !ok {
		return errors.Errorf("couldn't find plugin of type \"%s\" in the pluginlist", name)
	}

	inter := reflect.New(plug.Elem())
	s.JobStepper = inter.Interface().(JobStepper)

	// shove the data inside the config into the plugin
	data := parsed.Search("plugin_data")
	if data == nil {
		return errors.Errorf("no \"plugin_data\" blob associated with plugin \"%s\"", name)
	}

	err = json.Unmarshal(data.Bytes(), s.JobStepper)
	if err != nil {
		return errors.Wrapf(err, "couldn't Unmarshal \"plugin_data\" blob for plugin %s", name)
	}

	return nil
}
