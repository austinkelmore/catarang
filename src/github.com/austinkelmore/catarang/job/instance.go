package job

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/austinkelmore/catarang/cmd"
	"github.com/austinkelmore/catarang/plugin"
	"github.com/austinkelmore/catarang/template"
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
type InstJobStep struct {
	Log    cmd.Log
	Action plugin.JobStep
}

// Instance is a single run of a job
type Instance struct {
	StartTime time.Time
	EndTime   time.Time

	Template template.Job

	Steps []InstJobStep

	Status Status
	Error  error
}

// NewInstance creates a new instance from a job template
func NewInstance(t template.Job) (*Instance, error) {
	i := Instance{Template: t}

	var err error
	if i.Steps, err = createStepsFromTemplate(t); err != nil {
		return &i, errors.Wrapf(err, "error creating steps from template")
	}
	return &i, nil
}

func createStepsFromTemplate(t template.Job) ([]InstJobStep, error) {
	s := []InstJobStep{}
	path, err := filepath.Abs(t.LocalPath)
	if err != nil {
		return s, errors.Wrapf(err, "can't get absolute path from \"%s\"", t.LocalPath)
	}

	for _, step := range t.Steps {
		plug, ok := plugin.GetAvailable()[step.PluginName]
		if !ok {
			return s, errors.Errorf("couldn't find plugin of type \"%s\" in the available map", step.PluginName)
		}

		val := reflect.New(plug.Elem())
		jobstep := val.Interface().(plugin.JobStep)

		err := json.Unmarshal(step.PluginData, jobstep)
		if err != nil {
			return s, errors.Wrapf(err, "couldn't Unmarshal \"plugin\" blob for plugin %s", step.PluginName)
		}

		instStep := InstJobStep{Action: jobstep}
		instStep.Log.Name = instStep.Action.GetName()
		instStep.Log.WorkingDir = path
		s = append(s, instStep)
	}
	return s, nil
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
			if scm, ok := i.Steps[index].Action.(plugin.SCM); ok {
				if err := scm.FirstTimeSetup(&i.Steps[index].Log); err != nil {
					i.Error = errors.Wrapf(err, "can't run first FirstTimeSetup on scm %v", scm.GetName())
					i.Status = FAILED
					return
				}
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
			if i.Steps, err = createStepsFromTemplate(i.Template); err != nil {
				i.Error = errors.Wrapf(err, "error creating steps from template")
				i.Status = FAILED
				return
			}
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

	t := template.Job{LocalPath: i.Template.LocalPath}

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
