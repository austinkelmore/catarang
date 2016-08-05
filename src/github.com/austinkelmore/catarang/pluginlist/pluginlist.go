package pluginlist

import (
	"reflect"

	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/plugin/scm"
)

func Plugins() map[string]reflect.Type {
	return plugins
}

var plugins map[string]reflect.Type

func init() {
	plugins = make(map[string]reflect.Type)
	addPlugin(&scm.Git{})
	addPlugin(&plugin.RunCommand{})
	addPlugin(&plugin.Artifact{})
}

func addPlugin(p plugin.Runner) {
	plugins[p.GetName()] = reflect.ValueOf(p).Type()
}
