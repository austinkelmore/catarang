package ulog

import (
	"strings"
	"testing"
)

func TestEcho(t *testing.T) {
	output := "worked"

	j := StepLog{}
	cmd := j.New("echo", output)
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
	if strings.Compare(cmd.Str, output+"\n") != 0 {
		t.Errorf("Echo didn't print out to log correctly. Should have been: %s, was: %s\n", output, cmd.Str)
	}
}
