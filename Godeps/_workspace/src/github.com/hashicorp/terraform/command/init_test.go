package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/remote"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/terraform"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/cli"
)

func TestInit(t *testing.T) {
	dir := tempDir(t)

	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{
		testFixturePath("init"),
		dir,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}

	if _, err := os.Stat(filepath.Join(dir, "hello.tf")); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestInit_cwd(t *testing.T) {
	dir := tempDir(t)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Change to the temporary directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{
		testFixturePath("init"),
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}

	if _, err := os.Stat("hello.tf"); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestInit_multipleArgs(t *testing.T) {
	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{
		"bad",
		"bad",
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: \n%s", ui.OutputWriter.String())
	}
}

func TestInit_noArgs(t *testing.T) {
	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: \n%s", ui.OutputWriter.String())
	}
}

// https://github.com/hashicorp/terraform/issues/518
func TestInit_dstInSrc(t *testing.T) {
	dir := tempDir(t)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Change to the temporary directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Chdir(cwd)

	if _, err := os.Create("issue518.tf"); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{
		".",
		"foo",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}

	if _, err := os.Stat(filepath.Join(dir, "foo", "issue518.tf")); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestInit_remoteState(t *testing.T) {
	tmp, cwd := testCwd(t)
	defer testFixCwd(t, tmp, cwd)

	s := terraform.NewState()
	conf, srv := testRemoteState(t, s, 200)
	defer srv.Close()

	ui := new(cli.MockUi)
	c := &InitCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(testProvider()),
			Ui:          ui,
		},
	}

	args := []string{
		"-backend", "http",
		"-address", conf.Config["address"],
		testFixturePath("init"),
		tmp,
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: \n%s", ui.ErrorWriter.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "hello.tf")); err != nil {
		t.Fatalf("err: %s", err)
	}

	path, _ := remote.HiddenStatePath()
	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing state")
	}
}
