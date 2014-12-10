package command

import (
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/cli"
)

func TestVersionCommand_implements(t *testing.T) {
	var _ cli.Command = &VersionCommand{}
}
