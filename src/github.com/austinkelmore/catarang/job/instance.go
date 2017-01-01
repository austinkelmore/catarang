package job

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/austinkelmore/catarang/jobdata"
	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/plugin/scm"
	"github.com/austinkelmore/catarang/pluginlist"
	"github.com/pkg/errors"
)

// Status The job instance's status
type Status int

const (
	// INITIALIZED The job instance is initialized, but hasn't started running yet
	INITIALIZED Status = iota
	// RUNNING The job instance is currently running
	RUNNING
	// FAILED The job instance failed
	FAILED
	// SUCCESSFUL The job instance was successful
	SUCCESSFUL
)

// todo: akelmore - get generate stringer working with Status instead of hard coding it
func (s Status) String() string {
	switch s {
	case INITIALIZED:
		return "Initialized"
	case RUNNING:
		return "Running"
	case FAILED:
		return "Failed"
	case SUCCESSFUL:
		return "Successful"
	default:
		return "Unknown, not in String() function"
	}
}

// InstJobStep is a distinct use of a plugin to do a single step or action within a job
// todo: akelmore - figure out why InstJobStep is different from Step
type InstJobStep struct {
	Log    cmd.Log
	Action plugin.JobStep
}

// Instance is a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time

	Template jobdata.JobTemplate

	Steps []InstJobStep

	Status Status
	Error  error
}

// todo: akelmore - rename from NewInstance
// todo: akelmore - handle an error being thrown somewhere in here and return it
func NewInstance(t jobdata.JobTemplate) *Instance {
	i := Instance{Template: t}

	i.Steps = createStepsFromTemplate(t)
	return &i
}

func createStepsFromTemplate(t jobdata.JobTemplate) []InstJobStep {
	path, _ := filepath.Abs(t.LocalPath)
	// todo: akelmore - handle the error
	// if err != nil {
	// 	i.Error = errors.Wrapf(err, "can't get absolute path from \"%s\"", i.LocalPath)
	// 	i.Status = FAILED
	// 	return nil
	// }

	s := []InstJobStep{}
	for _, step := range t.Steps {
		plug, ok := pluginlist.Plugins()[step.PluginName]
		if !ok {
			// todo: akelmore - handle the error
			// i.Error = errors.Errorf("couldn't find plugin of type \"%s\" in the pluginlist", step.PluginName)
			// i.Status = FAILED
			return s
		}

		val := reflect.New(plug.Elem())
		jobstep := val.Interface().(plugin.JobStep)

		err := json.Unmarshal(step.PluginData, jobstep)
		if err != nil {
			// todo: akelmore - handle the error
			// i.Error = errors.Wrapf(err, "couldn't Unmarshal \"plugin\" blob for plugin %s", step.PluginName)
			// i.Status = FAILED
			return s
		}

		instStep := InstJobStep{Action: jobstep}
		instStep.Log.Name = instStep.Action.GetName()
		instStep.Log.WorkingDir = path
		s = append(s, instStep)
	}
	return s
}

// Start is an entry point for the instance
func (i *Instance) Start(doFirstTimeSetup bool) {
	i.StartTime = time.Now()
	defer func() { i.EndTime = time.Now() }()
	i.Status = RUNNING

	if err := os.MkdirAll(i.Template.LocalPath, 0777); err != nil {
		i.Error = errors.Wrapf(err, "can't create directory for job at path \"%s\"", i.Template.LocalPath)
		i.Status = FAILED
		return
	}

	if doFirstTimeSetup {
		for index := range i.Steps {
			if scm, ok := i.Steps[index].Action.(scm.SCMer); ok {
				// todo: akelmore - catch error and handle
				scm.FirstTimeSetup(&i.Steps[index].Log)
			}
		}
	}

	for needsUpdate := true; needsUpdate; {
		var err error
		needsUpdate, err = i.updateTemplate()
		if err != nil {
			i.Error = errors.Wrapf(err, "error updating the job template")
			i.Status = FAILED
			return
		}
		if needsUpdate {
			i.Steps = createStepsFromTemplate(i.Template)
		}
	}

	for index := range i.Steps {
		if err := i.Steps[index].Action.Run(&i.Steps[index].Log); err != nil {
			i.Error = errors.Wrapf(err, "couldn't run step index %v with action name %s", index, i.Steps[index].Action.GetName())
			i.Status = FAILED
			return
		}
	}

	i.Status = SUCCESSFUL
}

// updateTemplate updates the configuration information for the job.
// This needs to be done every time in case it has changed from the previous run.
func (i *Instance) updateTemplate() (bool, error) {
	// todo: akelmore - make updating the config use a logging that can be passed back to the job
	file := filepath.Join(i.Template.LocalPath, ".catarang.json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return false, errors.Wrapf(err, "error reading in config file \"%s\"", file)
	}

	t := jobdata.JobTemplate{LocalPath: i.Template.LocalPath}

	if err = json.Unmarshal(data, &t); err != nil {
		return false, errors.Wrapf(err, "error unmarshaling json from \"%s\"", file)
	}

	// test whether we need to redo our setup or not
	if !reflect.DeepEqual(t, i.Template) {
		i.Template = t
		return true, nil
	}

	return false, nil
}
