package multilog

import "bytes"

// Log captures output and error output separately
type Log struct {
	Name string
	Out  bytes.Buffer
	Err  bytes.Buffer
}

// New creates a new MultiLog
func New(name string) Log {
	log := Log{Name: name}
	return log
}
