package command

import (
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/config/module"
)

func TestUiModuleStorage_impl(t *testing.T) {
	var _ module.Storage = new(uiModuleStorage)
}
