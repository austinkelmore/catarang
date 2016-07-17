package jobcommand

// Cmd The structure of the job (what to run)
type Cmd struct {
	ExecCommands []string
	// Artifacts    []Artifact
}
