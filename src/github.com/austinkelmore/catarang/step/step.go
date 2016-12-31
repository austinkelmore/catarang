package step

import (
	"encoding/json"
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
)

type Template struct {
	PluginName string `json:"plugin_name"`
	plugin     plugin.JobStep
}

func New(p plugin.JobStep) Template {
	t := Template{PluginName: p.GetName(), plugin: p}
	return t
}

func (s Template) Plugin() plugin.JobStep {
	return s.plugin
}

// UnmarshalJSON converts arbitrary JSON into Go objects based on the plugins that are known in pluginlist.
func (s *Template) UnmarshalJSON(b []byte) error {
	parsed, err := gabs.ParseJSON(b)
	if err != nil {
		return errors.Wrap(err, "error parsing JSON while unmarshaling it")
	}

	pluginNameJSON := parsed.Search("plugin_name")
	if pluginNameJSON.Data() == nil {
		return errors.New("couldn't find \"plugin_name\" in StepTemplate")
	}

	name, ok := pluginNameJSON.Data().(string)
	if !ok {
		return errors.New("\"plugin_name\" does not reference a string")
	}
	s.PluginName = name

	plug, ok := pluginlist.Plugins()[s.PluginName]
	if !ok {
		return errors.Errorf("couldn't find plugin of type \"%s\" in the pluginlist", s.PluginName)
	}

	inter := reflect.New(plug.Elem())
	s.plugin = inter.Interface().(plugin.JobStep)

	// shove the data inside the config into the plugin
	data := parsed.Search("plugin")
	if data == nil {
		return errors.Errorf("no \"plugin\" blob associated with plugin \"%s\"", s.PluginName)
	}

	bytes := data.Bytes()
	err = json.Unmarshal(bytes, s.plugin)
	if err != nil {
		return errors.Wrapf(err, "couldn't Unmarshal \"plugin\" blob for plugin %s", s.PluginName)
	}

	return nil
}