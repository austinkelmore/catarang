package ulog

import "testing"

func TestEcho(t *testing.T) {
	output := "worked"
	cmd := Cmd{}
	cmd.init("echo", output)
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
	if string(cmd.Bytes()) != output+"\n" {
		t.Errorf("Echo didn't print out to log correctly. Should have been: %s, was: %s\n", output, cmd.Bytes())
	}
}
