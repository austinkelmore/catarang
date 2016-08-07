package job

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
)

type Step struct {
	Action plugin.Runner
	Name   string
}

// todo: akelmore - return the correct errors instead of returning nil
// todo: akelmore - handle logging so it can be returned back to the job
func (s *Step) UnmarshalJSON(b []byte) error {
	// todo: akelmore - handle parse error
	parsed, _ := gabs.ParseJSON(b)
	plug := parsed.Search("plugin")
	if plug == nil {
		return errors.New("Couldn't find \"plugin\" in Step.")
	}

	plugName, ok := plug.Data().(string)
	if !ok {
		return errors.New("\"plugin\" does not reference a string.")
	}

	actionType, ok := pluginlist.Plugins()[plugName]
	if !ok {
		return errors.New(fmt.Sprintf("Couldn't find plugin of type \"%s\" in the pluginlist. Have you added it there?.", plugName))
	}

	inter := reflect.New(actionType.Elem())
	s.Action = inter.Interface().(plugin.Runner)

	// shove the data inside the config into the plugin
	data := parsed.Search("data")
	if data == nil {
		return errors.New(fmt.Sprintf("No \"data\" blob associated with plugin \"%s\".", plugName))
	}

	bytes := data.Bytes()
	err := json.Unmarshal(bytes, s.Action)
	if err != nil {
		return errors.Wrapf(err, "Couldn't Unmarshal \"data\" blob for plugin %s.", plugName)
	}

	return nil
}
