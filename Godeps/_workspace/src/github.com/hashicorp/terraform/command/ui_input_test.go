package command

import (
	"bytes"
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/terraform"
)

func TestUIInput_impl(t *testing.T) {
	var _ terraform.UIInput = new(UIInput)
}

func TestUIInputInput(t *testing.T) {
	i := &UIInput{
		Reader: bytes.NewBufferString("foo\n"),
		Writer: bytes.NewBuffer(nil),
	}

	v, err := i.Input(&terraform.InputOpts{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if v != "foo" {
		t.Fatalf("bad: %#v", v)
	}
}
