package template

import "github.com/austinkelmore/catarang/plugin"

type Job struct {
	Steps     []plugin.JobStep
	LocalPath string
}
