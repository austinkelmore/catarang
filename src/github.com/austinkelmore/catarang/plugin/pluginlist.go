package plugin

import "reflect"

// GetAvailable returns the map of plugins that the system knows about
func GetAvailable() map[string]reflect.Type {
	return plugins
}

var plugins map[string]reflect.Type

func pluginList() []JobStep {
	// add all known plugins into this array
	pluginTypes := []JobStep{
		&Git{},
		&RunCommand{},
		&Artifact{},
	}
	return pluginTypes
}

func init() {
	plugins = make(map[string]reflect.Type)
	for _, plugin := range pluginList() {
		addPlugin(plugin)
	}
}

func addPlugin(p JobStep) {
	plugins[p.GetName()] = reflect.ValueOf(p).Type()
}
