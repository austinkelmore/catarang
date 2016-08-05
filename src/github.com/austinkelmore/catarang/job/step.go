package job

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/jeffail/gabs"
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
		log.Println("Couldn't find \"plugin\" in Step.")
		return nil
	}

	plugName, ok := plug.Data().(string)
	if !ok {
		log.Println("\"plugin\" was not a string in the config file.")
	}

	actionType, ok := pluginlist.Plugins()[plugName]
	if !ok {
		log.Printf("Couldn't find plugin of type \"%s\".\n", plugName)
		return nil
	}

	inter := reflect.New(actionType.Elem())
	s.Action = inter.Interface().(plugin.Runner)

	// shove the data inside the config into the plugin
	data := parsed.Search("data")
	if data == nil {
		log.Printf("No data blob in config associated with plugin \"%s\".\n", plugName)
		return nil
	}

	bytes := data.Bytes()
	err := json.Unmarshal(bytes, s.Action)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return nil
	}

	return nil
}
