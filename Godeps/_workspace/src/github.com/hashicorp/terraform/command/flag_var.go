package command

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/hcl"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/go-homedir"
)

// FlagVar is a flag.Value implementation for parsing user variables
// from the command-line in the format of '-var key=value'.
type FlagVar map[string]string

func (v *FlagVar) String() string {
	return ""
}

func (v *FlagVar) Set(raw string) error {
	idx := strings.Index(raw, "=")
	if idx == -1 {
		return fmt.Errorf("No '=' value in arg: %s", raw)
	}

	if *v == nil {
		*v = make(map[string]string)
	}

	key, value := raw[0:idx], raw[idx+1:]
	(*v)[key] = value
	return nil
}

// FlagVarFile is a flag.Value implementation for parsing user variables
// from the command line in the form of files. i.e. '-var-file=foo'
type FlagVarFile map[string]string

func (v *FlagVarFile) String() string {
	return ""
}

func (v *FlagVarFile) Set(raw string) error {
	vs, err := loadVarFile(raw)
	if err != nil {
		return err
	}

	if *v == nil {
		*v = make(map[string]string)
	}

	for key, value := range vs {
		(*v)[key] = value
	}

	return nil
}

func loadVarFile(rawPath string) (map[string]string, error) {
	path, err := homedir.Expand(rawPath)
	if err != nil {
		return nil, fmt.Errorf(
			"Error expanding path: %s", err)
	}

	// Read the HCL file and prepare for parsing
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"Error reading %s: %s", path, err)
	}

	// Parse it
	obj, err := hcl.Parse(string(d))
	if err != nil {
		return nil, fmt.Errorf(
			"Error parsing %s: %s", path, err)
	}

	var result map[string]string
	if err := hcl.DecodeObject(&result, obj); err != nil {
		return nil, err
	}

	return result, nil
}
