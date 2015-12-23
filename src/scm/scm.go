package scm

import "multilog"

// SCMer interface for the source control modules
type SCMer interface {
	FirstTimeSetup(logger *multilog.Log) error
	Poll(logger *multilog.Log) (bool, error)
	UpdateExisting(logger *multilog.Log) error
	LocalRepoPath() string
}
