package command

import (
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/terraform"
)

func TestCountHook_impl(t *testing.T) {
	var _ terraform.Hook = new(CountHook)
}
