package command

import (
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/cli"
)

func TestColorizeUi_impl(t *testing.T) {
	var _ cli.Ui = new(ColorizeUi)
}
