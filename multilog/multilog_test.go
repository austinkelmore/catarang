package multilog_test

import (
	"testing"

	"github.com/austinkelmore/catarang/multilog"
)

func TestNew(t *testing.T) {
	name := "randomstring"
	logger := multilog.New(name)
	if logger.Name != name {
		t.Error("Name wasn't saved in multilogger from constructor.")
	}
}
