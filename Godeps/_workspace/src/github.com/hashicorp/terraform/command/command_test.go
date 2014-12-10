package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/config/module"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/terraform"
)

// This is the directory where our test fixtures are.
var fixtureDir = "./test-fixtures"

func init() {
	test = true

	// Expand the fixture dir on init because we change the working
	// directory in some tests.
	var err error
	fixtureDir, err = filepath.Abs(fixtureDir)
	if err != nil {
		panic(err)
	}
}

func tempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	return dir
}

func testFixturePath(name string) string {
	return filepath.Join(fixtureDir, name)
}

func testCtxConfig(p terraform.ResourceProvider) *terraform.ContextOpts {
	return &terraform.ContextOpts{
		Providers: map[string]terraform.ResourceProviderFactory{
			"test": func() (terraform.ResourceProvider, error) {
				return p, nil
			},
		},
	}
}

func testModule(t *testing.T, name string) *module.Tree {
	mod, err := module.NewTreeModule("", filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	s := &module.FolderStorage{StorageDir: tempDir(t)}
	if err := mod.Load(s, module.GetModeGet); err != nil {
		t.Fatalf("err: %s", err)
	}

	return mod
}

func testPlanFile(t *testing.T, plan *terraform.Plan) string {
	path := testTempFile(t)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	if err := terraform.WritePlan(plan, f); err != nil {
		t.Fatalf("err: %s", err)
	}

	return path
}

func testReadPlan(t *testing.T, path string) *terraform.Plan {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	p, err := terraform.ReadPlan(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return p
}

// testState returns a test State structure that we use for a lot of tests.
func testState() *terraform.State {
	return &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}
}

func testStateFile(t *testing.T, s *terraform.State) string {
	path := testTempFile(t)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer f.Close()

	if err := terraform.WriteState(s, f); err != nil {
		t.Fatalf("err: %s", err)
	}

	return path
}

func testProvider() *terraform.MockResourceProvider {
	p := new(terraform.MockResourceProvider)
	p.DiffReturn = &terraform.InstanceDiff{}
	p.RefreshFn = func(
		info *terraform.InstanceInfo,
		s *terraform.InstanceState) (*terraform.InstanceState, error) {
		return s, nil
	}
	p.ResourcesReturn = []terraform.ResourceType{
		terraform.ResourceType{
			Name: "test_instance",
		},
	}

	return p
}

func testTempFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result := tf.Name()

	if err := tf.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.Remove(result); err != nil {
		t.Fatalf("err: %s", err)
	}

	return result
}

func testTempDir(t *testing.T) string {
	d, err := ioutil.TempDir("", "tf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return d
}

// testCwdDir is used to change the current working directory
// into a test directory that should be remoted after
func testCwd(t *testing.T) (string, string) {
	tmp, err := ioutil.TempDir("", "remote")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("err: %v", err)
	}
	return tmp, cwd
}

// testFixCwd is used to as a defer to testDir
func testFixCwd(t *testing.T, tmp, cwd string) {
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("err: %v", err)
	}
}
