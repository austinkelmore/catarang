package scm

import "io"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(outWriter io.Writer, errWriter io.Writer) error
	Poll() (bool, error)
	UpdateExisting(outWriter *io.Writer, errWriter *io.Writer) error
	LocalRepoPath() string
}
