package job

import (
	"encoding/json"
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
)

// Step is a single step in a job that is defined by a name which looks itself up in the pluginlist
type Step struct {
	Action plugin.Runner
	Name   string
}

// UnmarshalJSON converts arbitrary JSON into Go objects based on the plugins that are known in pluginlist.
func (s *Step) UnmarshalJSON(b []byte) error {
	parsed, err := gabs.ParseJSON(b)
	if err != nil {
		return errors.Wrap(err, "error parsing JSON while unmarshaling it")
	}

	plug := parsed.Search("plugin")
	if plug == nil {
		return errors.New("couldn't find \"plugin\" in Step")
	}

	plugName, ok := plug.Data().(string)
	if !ok {
		return errors.New("\"plugin\" does not reference a string")
	}

	actionType, ok := pluginlist.Plugins()[plugName]
	if !ok {
		return errors.Errorf("couldn't find plugin of type \"%s\" in the pluginlist", plugName)
	}

	inter := reflect.New(actionType.Elem())
	s.Action = inter.Interface().(plugin.Runner)

	// shove the data inside the config into the plugin
	data := parsed.Search("data")
	if data == nil {
		return errors.Errorf("No \"data\" blob associated with plugin \"%s\".", plugName)
	}

	bytes := data.Bytes()
	err = json.Unmarshal(bytes, s.Action)
	if err != nil {
		return errors.Wrapf(err, "Couldn't Unmarshal \"data\" blob for plugin %s.", plugName)
	}

	return nil
}
