package pluginlist

import (
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/plugin/scm"
)

// Plugins returns the map of plugins that the system knows about
func Plugins() map[string]reflect.Type {
	return plugins
}

var plugins map[string]reflect.Type

func pluginList() []plugin.JobStep {
	// add all known plugins into this array
	pluginTypes := []plugin.JobStep{
		&scm.Git{},
		&plugin.RunCommand{},
		&plugin.Artifact{},
	}
	return pluginTypes
}

func init() {
	plugins = make(map[string]reflect.Type)
	for _, plugin := range pluginList() {
		addPlugin(plugin)
	}
}

func addPlugin(p plugin.JobStep) {
	plugins[p.GetName()] = reflect.ValueOf(p).Type()
}
