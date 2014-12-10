// The config package is responsible for loading and validating the
// configuration.
package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/flatmap"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/hashicorp/terraform/helper/multierror"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/mapstructure"
	"github.com/svend/terraform/Godeps/_workspace/src/github.com/mitchellh/reflectwalk"
)

// NameRegexp is the regular expression that all names (modules, providers,
// resources, etc.) must follow.
var NameRegexp = regexp.MustCompile(`\A[A-Za-z0-9\-\_]+\z`)

// Config is the configuration that comes from loading a collection
// of Terraform templates.
type Config struct {
	// Dir is the path to the directory where this configuration was
	// loaded from. If it is blank, this configuration wasn't loaded from
	// any meaningful directory.
	Dir string

	Modules         []*Module
	ProviderConfigs []*ProviderConfig
	Resources       []*Resource
	Variables       []*Variable
	Outputs         []*Output

	// The fields below can be filled in by loaders for validation
	// purposes.
	unknownKeys []string
}

// Module is a module used within a configuration.
//
// This does not represent a module itself, this represents a module
// call-site within an existing configuration.
type Module struct {
	Name      string
	Source    string
	RawConfig *RawConfig
}

// ProviderConfig is the configuration for a resource provider.
//
// For example, Terraform needs to set the AWS access keys for the AWS
// resource provider.
type ProviderConfig struct {
	Name      string
	RawConfig *RawConfig
}

// A resource represents a single Terraform resource in the configuration.
// A Terraform resource is something that represents some component that
// can be created and managed, and has some properties associated with it.
type Resource struct {
	Name         string
	Type         string
	RawCount     *RawConfig
	RawConfig    *RawConfig
	Provisioners []*Provisioner
	DependsOn    []string
	Lifecycle    ResourceLifecycle
}

// ResourceLifecycle is used to store the lifecycle tuning parameters
// to allow customized behavior
type ResourceLifecycle struct {
	CreateBeforeDestroy bool `hcl:"create_before_destroy"`
}

// Provisioner is a configured provisioner step on a resource.
type Provisioner struct {
	Type      string
	RawConfig *RawConfig
	ConnInfo  *RawConfig
}

// Variable is a variable defined within the configuration.
type Variable struct {
	Name        string
	Default     interface{}
	Description string
}

// Output is an output defined within the configuration. An output is
// resulting data that is highlighted by Terraform when finished.
type Output struct {
	Name      string
	RawConfig *RawConfig
}

// VariableType is the type of value a variable is holding, and returned
// by the Type() function on variables.
type VariableType byte

const (
	VariableTypeUnknown VariableType = iota
	VariableTypeString
	VariableTypeMap
)

// ProviderConfigName returns the name of the provider configuration in
// the given mapping that maps to the proper provider configuration
// for this resource.
func ProviderConfigName(t string, pcs []*ProviderConfig) string {
	lk := ""
	for _, v := range pcs {
		k := v.Name
		if strings.HasPrefix(t, k) && len(k) > len(lk) {
			lk = k
		}
	}

	return lk
}

// A unique identifier for this module.
func (r *Module) Id() string {
	return fmt.Sprintf("%s", r.Name)
}

// Count returns the count of this resource.
func (r *Resource) Count() (int, error) {
	v, err := strconv.ParseInt(r.RawCount.Value().(string), 0, 0)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

// A unique identifier for this resource.
func (r *Resource) Id() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}

// Validate does some basic semantic checking of the configuration.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, k := range c.unknownKeys {
		errs = append(errs, fmt.Errorf(
			"Unknown root level key: %s", k))
	}

	vars := c.InterpolatedVariables()
	varMap := make(map[string]*Variable)
	for _, v := range c.Variables {
		varMap[v.Name] = v
	}

	for _, v := range c.Variables {
		if v.Type() == VariableTypeUnknown {
			errs = append(errs, fmt.Errorf(
				"Variable '%s': must be string or mapping",
				v.Name))
			continue
		}

		interp := false
		fn := func(i Interpolation) (string, error) {
			interp = true
			return "", nil
		}

		w := &interpolationWalker{F: fn}
		if v.Default != nil {
			if err := reflectwalk.Walk(v.Default, w); err == nil {
				if interp {
					errs = append(errs, fmt.Errorf(
						"Variable '%s': cannot contain interpolations",
						v.Name))
				}
			}
		}
	}

	// Check for references to user variables that do not actually
	// exist and record those errors.
	for source, vs := range vars {
		for _, v := range vs {
			uv, ok := v.(*UserVariable)
			if !ok {
				continue
			}

			if _, ok := varMap[uv.Name]; !ok {
				errs = append(errs, fmt.Errorf(
					"%s: unknown variable referenced: %s",
					source,
					uv.Name))
			}
		}
	}

	// Check that all count variables are valid.
	for source, vs := range vars {
		for _, rawV := range vs {
			switch v := rawV.(type) {
			case *CountVariable:
				if v.Type == CountValueInvalid {
					errs = append(errs, fmt.Errorf(
						"%s: invalid count variable: %s",
						source,
						v.FullKey()))
				}
			case *PathVariable:
				if v.Type == PathValueInvalid {
					errs = append(errs, fmt.Errorf(
						"%s: invalid path variable: %s",
						source,
						v.FullKey()))
				}
			}
		}
	}

	// Check that all references to modules are valid
	modules := make(map[string]*Module)
	dupped := make(map[string]struct{})
	for _, m := range c.Modules {
		// Check for duplicates
		if _, ok := modules[m.Id()]; ok {
			if _, ok := dupped[m.Id()]; !ok {
				dupped[m.Id()] = struct{}{}

				errs = append(errs, fmt.Errorf(
					"%s: module repeated multiple times",
					m.Id()))
			}
		}

		if _, ok := modules[m.Id()]; !ok {
			// If we haven't seen this module before, check that the
			// source has no interpolations.
			rc, err := NewRawConfig(map[string]interface{}{
				"root": m.Source,
			})
			if err != nil {
				errs = append(errs, fmt.Errorf(
					"%s: module source error: %s",
					m.Id(), err))
			} else if len(rc.Interpolations) > 0 {
				errs = append(errs, fmt.Errorf(
					"%s: module source cannot contain interpolations",
					m.Id()))
			}

			// Check that the name matches our regexp
			if !NameRegexp.Match([]byte(m.Name)) {
				errs = append(errs, fmt.Errorf(
					"%s: module name can only contain letters, numbers, "+
						"dashes, and underscores",
					m.Id()))
			}
		}

		modules[m.Id()] = m
	}
	dupped = nil

	// Check that all variables for modules reference modules that
	// exist.
	for source, vs := range vars {
		for _, v := range vs {
			mv, ok := v.(*ModuleVariable)
			if !ok {
				continue
			}

			if _, ok := modules[mv.Name]; !ok {
				errs = append(errs, fmt.Errorf(
					"%s: unknown module referenced: %s",
					source,
					mv.Name))
			}
		}
	}

	// Check that all references to resources are valid
	resources := make(map[string]*Resource)
	dupped = make(map[string]struct{})
	for _, r := range c.Resources {
		if _, ok := resources[r.Id()]; ok {
			if _, ok := dupped[r.Id()]; !ok {
				dupped[r.Id()] = struct{}{}

				errs = append(errs, fmt.Errorf(
					"%s: resource repeated multiple times",
					r.Id()))
			}
		}

		resources[r.Id()] = r
	}
	dupped = nil

	// Validate resources
	for n, r := range resources {
		// Verify count variables
		for _, v := range r.RawCount.Variables {
			switch v.(type) {
			case *CountVariable:
				errs = append(errs, fmt.Errorf(
					"%s: resource count can't reference count variable: %s",
					n,
					v.FullKey()))
			case *ModuleVariable:
				errs = append(errs, fmt.Errorf(
					"%s: resource count can't reference module variable: %s",
					n,
					v.FullKey()))
			case *ResourceVariable:
				errs = append(errs, fmt.Errorf(
					"%s: resource count can't reference resource variable: %s",
					n,
					v.FullKey()))
			case *UserVariable:
				// Good
			default:
				panic("Unknown type in count var: " + n)
			}
		}

		// Interpolate with a fixed number to verify that its a number
		r.RawCount.interpolate(func(Interpolation) (string, error) {
			return "5", nil
		})
		_, err := strconv.ParseInt(r.RawCount.Value().(string), 0, 0)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"%s: resource count must be an integer",
				n))
		}
		r.RawCount.init()

		for _, d := range r.DependsOn {
			if _, ok := resources[d]; !ok {
				errs = append(errs, fmt.Errorf(
					"%s: resource depends on non-existent resource '%s'",
					n, d))
			}
		}
	}

	for source, vs := range vars {
		for _, v := range vs {
			rv, ok := v.(*ResourceVariable)
			if !ok {
				continue
			}

			id := fmt.Sprintf("%s.%s", rv.Type, rv.Name)
			if _, ok := resources[id]; !ok {
				errs = append(errs, fmt.Errorf(
					"%s: unknown resource '%s' referenced in variable %s",
					source,
					id,
					rv.FullKey()))
				continue
			}
		}
	}

	// Check that all outputs are valid
	for _, o := range c.Outputs {
		invalid := false
		for k, _ := range o.RawConfig.Raw {
			if k != "value" {
				invalid = true
				break
			}
		}
		if invalid {
			errs = append(errs, fmt.Errorf(
				"%s: output should only have 'value' field", o.Name))
		}
	}

	// Check that all variables are in the proper context
	for source, rc := range c.rawConfigs() {
		walker := &interpolationWalker{
			ContextF: c.validateVarContextFn(source, &errs),
		}
		if err := reflectwalk.Walk(rc.Raw, walker); err != nil {
			errs = append(errs, fmt.Errorf(
				"%s: error reading config: %s", source, err))
		}
	}

	if len(errs) > 0 {
		return &multierror.Error{Errors: errs}
	}

	return nil
}

// InterpolatedVariables is a helper that returns a mapping of all the interpolated
// variables within the configuration. This is used to verify references
// are valid in the Validate step.
func (c *Config) InterpolatedVariables() map[string][]InterpolatedVariable {
	result := make(map[string][]InterpolatedVariable)
	for source, rc := range c.rawConfigs() {
		for _, v := range rc.Variables {
			result[source] = append(result[source], v)
		}
	}
	return result
}

// rawConfigs returns all of the RawConfigs that are available keyed by
// a human-friendly source.
func (c *Config) rawConfigs() map[string]*RawConfig {
	result := make(map[string]*RawConfig)
	for _, pc := range c.ProviderConfigs {
		source := fmt.Sprintf("provider config '%s'", pc.Name)
		result[source] = pc.RawConfig
	}

	for _, rc := range c.Resources {
		source := fmt.Sprintf("resource '%s'", rc.Id())
		result[source+" count"] = rc.RawCount
		result[source+" config"] = rc.RawConfig
	}

	for _, o := range c.Outputs {
		source := fmt.Sprintf("output '%s'", o.Name)
		result[source] = o.RawConfig
	}

	return result
}

func (c *Config) validateVarContextFn(
	source string, errs *[]error) interpolationWalkerContextFunc {
	return func(loc reflectwalk.Location, i Interpolation) {
		vi, ok := i.(*VariableInterpolation)
		if !ok {
			return
		}

		rv, ok := vi.Variable.(*ResourceVariable)
		if !ok {
			return
		}

		if rv.Multi && rv.Index == -1 && loc != reflectwalk.SliceElem {
			*errs = append(*errs, fmt.Errorf(
				"%s: multi-variable must be in a slice", source))
		}
	}
}

func (m *Module) mergerName() string {
	return m.Id()
}

func (m *Module) mergerMerge(other merger) merger {
	m2 := other.(*Module)

	result := *m
	result.Name = m2.Name
	result.RawConfig = result.RawConfig.merge(m2.RawConfig)

	if m2.Source != "" {
		result.Source = m2.Source
	}

	return &result
}

func (o *Output) mergerName() string {
	return o.Name
}

func (o *Output) mergerMerge(m merger) merger {
	o2 := m.(*Output)

	result := *o
	result.Name = o2.Name
	result.RawConfig = result.RawConfig.merge(o2.RawConfig)

	return &result
}

func (c *ProviderConfig) mergerName() string {
	return c.Name
}

func (c *ProviderConfig) mergerMerge(m merger) merger {
	c2 := m.(*ProviderConfig)

	result := *c
	result.Name = c2.Name
	result.RawConfig = result.RawConfig.merge(c2.RawConfig)

	return &result
}

func (r *Resource) mergerName() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}

func (r *Resource) mergerMerge(m merger) merger {
	r2 := m.(*Resource)

	result := *r
	result.Name = r2.Name
	result.Type = r2.Type
	result.RawConfig = result.RawConfig.merge(r2.RawConfig)

	if r2.RawCount.Value() != "1" {
		result.RawCount = r2.RawCount
	}

	if len(r2.Provisioners) > 0 {
		result.Provisioners = r2.Provisioners
	}

	return &result
}

// DefaultsMap returns a map of default values for this variable.
func (v *Variable) DefaultsMap() map[string]string {
	if v.Default == nil {
		return nil
	}

	n := fmt.Sprintf("var.%s", v.Name)
	switch v.Type() {
	case VariableTypeString:
		return map[string]string{n: v.Default.(string)}
	case VariableTypeMap:
		result := flatmap.Flatten(map[string]interface{}{
			n: v.Default.(map[string]string),
		})
		result[n] = v.Name

		return result
	default:
		return nil
	}
}

// Merge merges two variables to create a new third variable.
func (v *Variable) Merge(v2 *Variable) *Variable {
	// Shallow copy the variable
	result := *v

	// The names should be the same, but the second name always wins.
	result.Name = v2.Name

	if v2.Default != nil {
		result.Default = v2.Default
	}
	if v2.Description != "" {
		result.Description = v2.Description
	}

	return &result
}

// Type returns the type of varialbe this is.
func (v *Variable) Type() VariableType {
	if v.Default == nil {
		return VariableTypeString
	}

	var strVal string
	if err := mapstructure.WeakDecode(v.Default, &strVal); err == nil {
		v.Default = strVal
		return VariableTypeString
	}

	var m map[string]string
	if err := mapstructure.WeakDecode(v.Default, &m); err == nil {
		v.Default = m
		return VariableTypeMap
	}

	return VariableTypeUnknown
}

func (v *Variable) mergerName() string {
	return v.Name
}

func (v *Variable) mergerMerge(m merger) merger {
	return v.Merge(m.(*Variable))
}

// Required tests whether a variable is required or not.
func (v *Variable) Required() bool {
	return v.Default == nil
}
