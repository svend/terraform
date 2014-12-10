package terraform

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestGraph_basic(t *testing.T) {
	m := testModule(t, "graph-basic")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_configRequired(t *testing.T) {
	if _, err := Graph(new(GraphOpts)); err == nil {
		t.Fatal("should error")
	}
}

func TestGraph_count(t *testing.T) {
	m := testModule(t, "graph-count")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphCountStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_countTainted(t *testing.T) {
	m := testModule(t, "graph-count")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: []string{"root"},
				Resources: map[string]*ResourceState{
					"aws_instance.web.0": &ResourceState{
						Type: "aws_instance",
						Tainted: []*InstanceState{
							&InstanceState{
								ID: "foo",
							},
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphCountTaintedStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_varResource(t *testing.T) {
	m := testModule(t, "graph-count-var-resource")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphCountVarResourceStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_cycle(t *testing.T) {
	m := testModule(t, "graph-cycle")

	_, err := Graph(&GraphOpts{Module: m})
	if err == nil {
		t.Fatal("should error")
	}
}

func TestGraph_dependsOn(t *testing.T) {
	m := testModule(t, "graph-depends-on")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDependsStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_dependsOnCount(t *testing.T) {
	m := testModule(t, "graph-depends-on-count")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDependsCountStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_dependsOnWithOrphan(t *testing.T) {
	m := testModule(t, "graph-depends-on")

	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: []string{"root"},
				Resources: map[string]*ResourceState{
					"aws_instance.old": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDependsOrphanStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_modules(t *testing.T) {
	m := testModule(t, "graph-modules")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphModulesStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	n := g.Noun("module.consul")
	if n == nil {
		t.Fatal("can't find noun")
	}
	mn := n.Meta.(*GraphNodeModule)

	if !reflect.DeepEqual(mn.Path, []string{"root", "consul"}) {
		t.Fatalf("bad: %#v", mn.Path)
	}

	actual = strings.TrimSpace(mn.Graph.String())
	expected = strings.TrimSpace(testTerraformGraphModulesConsulStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_moduleOrphan(t *testing.T) {
	m := testModule(t, "graph-module-orphan")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: []string{"root", "consul"},

				Resources: map[string]*ResourceState{
					"aws_instance.old": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphModuleOrphanStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	n := g.Noun("module.consul")
	if n == nil {
		t.Fatal("can't find noun")
	}
	mn := n.Meta.(*GraphNodeModule)

	if !reflect.DeepEqual(mn.Path, []string{"root", "consul"}) {
		t.Fatalf("bad: %#v", mn.Path)
	}

	actual = strings.TrimSpace(mn.Graph.String())
	expected = strings.TrimSpace(testTerraformGraphModuleOrphanConsulStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_providerPrune(t *testing.T) {
	m := testModule(t, "graph-provider-prune")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphProviderPruneStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_state(t *testing.T) {
	m := testModule(t, "graph-basic")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.old": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphStateStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_tainted(t *testing.T) {
	m := testModule(t, "graph-tainted")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
						Tainted: []*InstanceState{
							&InstanceState{
								ID: "bar",
							},
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphTaintedStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraph_taintedMulti(t *testing.T) {
	m := testModule(t, "graph-tainted")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.web": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
						Tainted: []*InstanceState{
							&InstanceState{
								ID: "bar",
							},
							&InstanceState{
								ID: "baz",
							},
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphTaintedMultiStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraphFull(t *testing.T) {
	rpAws := new(MockResourceProvider)
	rpOS := new(MockResourceProvider)

	rpAws.ResourcesReturn = []ResourceType{
		ResourceType{Name: "aws_instance"},
		ResourceType{Name: "aws_load_balancer"},
		ResourceType{Name: "aws_security_group"},
	}
	rpOS.ResourcesReturn = []ResourceType{
		ResourceType{Name: "openstack_floating_ip"},
	}

	ps := map[string]ResourceProviderFactory{
		"aws":  testProviderFuncFixed(rpAws),
		"open": testProviderFuncFixed(rpOS),
	}

	m := testModule(t, "graph-basic")
	g, err := Graph(&GraphOpts{Module: m, Providers: ps})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// A helper to help get us the provider for a resource.
	graphProvider := func(n string) ResourceProvider {
		return g.Noun(n).Meta.(*GraphNodeResource).Resource.Provider
	}

	// Test a couple
	if graphProvider("aws_instance.web") != rpAws {
		t.Fatalf("bad: %#v", graphProvider("aws_instance.web"))
	}
	if graphProvider("openstack_floating_ip.random") != rpOS {
		t.Fatalf("bad: %#v", graphProvider("openstack_floating_ip.random"))
	}

	// Test that all providers have been set
	for _, n := range g.Nouns {
		switch m := n.Meta.(type) {
		case *GraphNodeResource:
			if m.Resource.Provider == nil {
				t.Fatalf("bad: %#v", m)
			}
		case *GraphNodeResourceProvider:
			if len(m.Provider.Providers) == 0 {
				t.Fatalf("bad: %#v", m)
			}
		default:
			continue
		}
	}
}

func TestGraphProvisioners(t *testing.T) {
	rpAws := new(MockResourceProvider)
	provShell := new(MockResourceProvisioner)
	provWinRM := new(MockResourceProvisioner)

	rpAws.ResourcesReturn = []ResourceType{
		ResourceType{Name: "aws_instance"},
		ResourceType{Name: "aws_load_balancer"},
		ResourceType{Name: "aws_security_group"},
	}

	ps := map[string]ResourceProvisionerFactory{
		"shell": testProvisionerFuncFixed(provShell),
		"winrm": testProvisionerFuncFixed(provWinRM),
	}

	pf := map[string]ResourceProviderFactory{
		"aws": testProviderFuncFixed(rpAws),
	}

	m := testModule(t, "graph-provisioners")
	g, err := Graph(&GraphOpts{Module: m, Providers: pf, Provisioners: ps})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// A helper to help get us the provider for a resource.
	graphProvisioner := func(n string, idx int) *ResourceProvisionerConfig {
		return g.Noun(n).Meta.(*GraphNodeResource).Resource.Provisioners[idx]
	}

	// A helper to verify depedencies
	depends := func(a, b string) bool {
		aNoun := g.Noun(a)
		bNoun := g.Noun(b)
		for _, dep := range aNoun.Deps {
			if dep.Source == aNoun && dep.Target == bNoun {
				return true
			}
		}
		return false
	}

	// Test a couple
	prov := graphProvisioner("aws_instance.web", 0)
	if prov.Provisioner != provWinRM {
		t.Fatalf("bad: %#v", prov)
	}
	if prov.RawConfig.Config()["cmd"] != "echo foo" {
		t.Fatalf("bad: %#v", prov)
	}

	prov = graphProvisioner("aws_instance.web", 1)
	if prov.Provisioner != provWinRM {
		t.Fatalf("bad: %#v", prov)
	}
	if prov.RawConfig.Config()["cmd"] != "echo bar" {
		t.Fatalf("bad: %#v", prov)
	}

	prov = graphProvisioner("aws_load_balancer.weblb", 0)
	if prov.Provisioner != provShell {
		t.Fatalf("bad: %#v", prov)
	}
	if prov.RawConfig.Config()["cmd"] != "add ${aws_instance.web.id}" {
		t.Fatalf("bad: %#v", prov)
	}
	if prov.ConnInfo == nil || len(prov.ConnInfo.Raw) != 2 {
		t.Fatalf("bad: %#v", prov)
	}

	// Check that the variable dependency is handled
	if !depends("aws_load_balancer.weblb", "aws_instance.web") {
		t.Fatalf("missing dependency from provisioner variable")
	}

	// Check that the connection variable dependency is handled
	if !depends("aws_load_balancer.weblb", "aws_security_group.firewall") {
		t.Fatalf("missing dependency from provisioner connection")
	}
}

func TestGraphAddDiff(t *testing.T) {
	m := testModule(t, "graph-diff")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Attributes: map[string]*ResourceAttrDiff{
							"foo": &ResourceAttrDiff{
								New: "bar",
							},
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, Diff: diff})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}

	/*
		TODO: test this somewhere
		// Verify that the state has been added
		n := g.Noun("aws_instance.foo")
		rn := n.Meta.(*GraphNodeResource)

		expected2 := diff.RootModule().Resources["aws_instance.foo"]
		actual2 := rn.Resource.Diff
		if !reflect.DeepEqual(actual2, expected2) {
			t.Fatalf("bad: %#v", actual2)
		}
	*/
}

func TestGraphAddDiff_destroy(t *testing.T) {
	m := testModule(t, "graph-diff-destroy")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
					"aws_instance.bar": &InstanceDiff{
						Destroy: true,
					},
				},
			},
		},
	}
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.foo": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},

					"aws_instance.bar": &ResourceState{
						Type:         "aws_instance",
						Dependencies: []string{"foo"},
						Primary: &InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	diffHash := checksumStruct(t, diff)

	g, err := Graph(&GraphOpts{
		Module: m,
		Diff:   diff,
		State:  state,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffDestroyStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s\n\nexpected:\n\n%s", actual, expected)
	}

	// Verify that the state has been added
	n := g.Noun("aws_instance.foo (destroy)")
	rn := n.Meta.(*GraphNodeResource)

	expected2 := &InstanceDiff{Destroy: true}
	actual2 := rn.Resource.Diff
	if !reflect.DeepEqual(actual2, expected2) {
		t.Fatalf("bad: %#v", actual2)
	}

	// Verify that our original structure has not been modified
	diffHash2 := checksumStruct(t, diff)
	if diffHash != diffHash2 {
		t.Fatal("diff has been modified")
	}
}

func TestGraphAddDiff_destroy_counts(t *testing.T) {
	m := testModule(t, "graph-count")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.web.0": &InstanceDiff{
						Destroy: true,
					},
					"aws_instance.web.1": &InstanceDiff{
						Destroy: true,
					},
					"aws_instance.web.2": &InstanceDiff{
						Destroy: true,
					},
					"aws_load_balancer.weblb": &InstanceDiff{
						Destroy: true,
					},
				},
			},
		},
	}
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web.0": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_instance.web.1": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_instance.web.2": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_load_balancer.weblb": &ResourceState{
						Type:         "aws_load_balancer",
						Dependencies: []string{"aws_instance.web.0", "aws_instance.web.1", "aws_instance.web.2"},
						Primary: &InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	diffHash := checksumStruct(t, diff)

	g, err := Graph(&GraphOpts{
		Module: m,
		Diff:   diff,
		State:  state,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffDestroyCountsStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s\n\nexpected:\n\n%s", actual, expected)
	}

	// Verify that the state has been added
	n := g.Noun("aws_instance.web (destroy)")
	rn := n.Meta.(*GraphNodeResource)

	if rn.ExpandMode != ResourceExpandDestroy {
		t.Fatalf("bad: %#v", rn)
	}

	// Verify that our original structure has not been modified
	diffHash2 := checksumStruct(t, diff)
	if diffHash != diffHash2 {
		t.Fatal("diff has been modified")
	}
}

func TestGraphAddDiff_module(t *testing.T) {
	m := testModule(t, "graph-diff-module")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, Diff: diff})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffModuleStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraphAddDiff_module_depends(t *testing.T) {
	m := testModule(t, "graph-diff-module-dep")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
				},
			},
			&ModuleDiff{
				Path:    []string{"root", "child"},
				Destroy: true,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
				},
			},
		},
	}
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: []string{"root", "orphan"},
				Resources: map[string]*ResourceState{
					"aws_instance.dead": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "dead",
						},
					},
				},
				Dependencies: []string{
					"aws_instance.foo",
					"module.child",
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, Diff: diff, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffModuleDependsStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraphAddDiff_createBeforeDestroy(t *testing.T) {
	m := testModule(t, "graph-diff-create-before")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.bar": &InstanceDiff{
						Destroy: true,
						Attributes: map[string]*ResourceAttrDiff{
							"ami": &ResourceAttrDiff{
								Old:         "abc",
								New:         "xyz",
								RequiresNew: true,
							},
						},
					},
				},
			},
		},
	}
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.bar": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"ami": "abc",
							},
						},
					},
				},
			},
		},
	}

	diffHash := checksumStruct(t, diff)

	g, err := Graph(&GraphOpts{
		Module: m,
		Diff:   diff,
		State:  state,
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffCreateBeforeDestroyStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s\n\nexpected:\n\n%s", actual, expected)
	}

	// Verify the flags are set
	r := g.Noun("aws_instance.bar")
	if r.Meta.(*GraphNodeResource).Resource.Flags&FlagReplacePrimary == 0 {
		t.Fatalf("missing FlagReplacePrimary")
	}

	r = g.Noun("aws_instance.bar (destroy)")
	if r.Meta.(*GraphNodeResource).Resource.Flags&FlagDeposed == 0 {
		t.Fatalf("missing FlagDeposed")
	}

	// Verify that our original structure has not been modified
	diffHash2 := checksumStruct(t, diff)
	if diffHash != diffHash2 {
		t.Fatal("diff has been modified")
	}
}

func TestGraphAddDiff_moduleDestroy(t *testing.T) {
	m := testModule(t, "graph-diff-module")
	diff := &Diff{
		Modules: []*ModuleDiff{
			&ModuleDiff{
				Path: rootModulePath,
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
				},
			},
			&ModuleDiff{
				Path: []string{"root", "child"},
				Resources: map[string]*InstanceDiff{
					"aws_instance.foo": &InstanceDiff{
						Destroy: true,
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, Diff: diff})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphDiffModuleStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestGraphEncodeDependencies(t *testing.T) {
	m := testModule(t, "graph-basic")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// This should encode the dependency information into the state
	graphEncodeDependencies(g)

	web := g.Noun("aws_instance.web").Meta.(*GraphNodeResource).Resource
	if len(web.Dependencies) != 1 || web.Dependencies[0] != "aws_security_group.firewall" {
		t.Fatalf("bad: %#v", web)
	}

	weblb := g.Noun("aws_load_balancer.weblb").Meta.(*GraphNodeResource).Resource
	if len(weblb.Dependencies) != 1 || weblb.Dependencies[0] != "aws_instance.web" {
		t.Fatalf("bad: %#v", weblb)
	}
}

func TestGraphEncodeDependencies_count(t *testing.T) {
	m := testModule(t, "graph-count")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,
				Resources: map[string]*ResourceState{
					"aws_instance.web.0": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_load_balancer.weblb": &ResourceState{
						Type: "aws_load_balancer",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// This should encode the dependency information into the state
	graphEncodeDependencies(g)

	web := g.Noun("aws_instance.web").Meta.(*GraphNodeResource).Resource
	if len(web.Dependencies) != 0 {
		t.Fatalf("bad: %#v", web)
	}

	weblb := g.Noun("aws_load_balancer.weblb").Meta.(*GraphNodeResource).Resource
	if len(weblb.Dependencies) != 1 {
		t.Fatalf("bad: %#v", weblb)
	}
}

func TestGraphEncodeDependencies_module(t *testing.T) {
	m := testModule(t, "graph-modules")

	g, err := Graph(&GraphOpts{Module: m, State: &State{}})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// This should encode the dependency information into the state
	graphEncodeDependencies(g)

	web := g.Noun("aws_instance.web").Meta.(*GraphNodeResource).Resource
	sort.Strings(web.Dependencies)
	if len(web.Dependencies) != 2 {
		t.Fatalf("bad: %#v", web)
	}
	if web.Dependencies[0] != "aws_security_group.firewall" {
		t.Fatalf("bad: %#v", web)
	}
	if web.Dependencies[1] != "module.consul" {
		t.Fatalf("bad: %#v", web)
	}

	mod := g.Noun("module.consul").Meta.(*GraphNodeModule)
	deps := mod.State.Dependencies
	if len(deps) != 1 {
		t.Fatalf("Bad: %#v", deps)
	}
	if deps[0] != "aws_security_group.firewall" {
		t.Fatalf("Bad: %#v", deps)
	}
}

func TestGraph_orphan_dependencies(t *testing.T) {
	m := testModule(t, "graph-count")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.web.0": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_instance.web.1": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
					},
					"aws_load_balancer.old": &ResourceState{
						Type: "aws_load_balancer",
						Primary: &InstanceState{
							ID: "foo",
						},
						Dependencies: []string{
							"aws_instance.web.0",
							"aws_instance.web.1",
							"aws_instance.web.2",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphCountOrphanStr)
	if actual != expected {
		t.Fatalf("bad:\n\nactual:\n%s\n\nexpected:\n%s", actual, expected)
	}
}

func TestGraph_orphanDependenciesModules(t *testing.T) {
	m := testModule(t, "graph-modules")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.foo": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
						Dependencies: []string{
							"module.consul",
						},
					},
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphOrphanModuleDepsStr)
	if actual != expected {
		t.Fatalf("bad:\n\nactual:\n%s\n\nexpected:\n%s", actual, expected)
	}
}

func TestGraph_orphanModules_Dependencies(t *testing.T) {
	m := testModule(t, "graph-modules")
	state := &State{
		Modules: []*ModuleState{
			&ModuleState{
				Path: rootModulePath,

				Resources: map[string]*ResourceState{
					"aws_instance.foo": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "foo",
						},
						Dependencies: []string{
							"module.consul",
						},
					},
				},
			},

			// Add an orphan module
			&ModuleState{
				Path: []string{"root", "orphan"},
				Resources: map[string]*ResourceState{
					"aws_instance.bar": &ResourceState{
						Type: "aws_instance",
						Primary: &InstanceState{
							ID: "bar",
						},
					},
				},
				Dependencies: []string{
					"aws_instance.foo",
					"aws_instance.web",
				},
			},
		},
	}

	g, err := Graph(&GraphOpts{Module: m, State: state})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphOrphanedModuleDepsStr)
	if actual != expected {
		t.Fatalf("bad:\n\nactual:\n%s\n\nexpected:\n%s", actual, expected)
	}
}

func TestGraphNodeResourceExpand(t *testing.T) {
	m := testModule(t, "graph-resource-expand")

	g, err := Graph(&GraphOpts{Module: m})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get the resource we care about expanding
	n := g.Noun("aws_instance.web")
	if n == nil {
		t.Fatal("could not find")
	}
	rn := n.Meta.(*GraphNodeResource)

	g, err = rn.Expand()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphResourceExpandStr)
	if actual != expected {
		t.Fatalf("bad:\n\nactual:\n%s\n\nexpected:\n%s", actual, expected)
	}
}

func TestGraphNodeResourceExpand_provDeps(t *testing.T) {
	m := testModule(t, "graph-resource-expand-prov-deps")
	provs := map[string]ResourceProvisionerFactory{
		"remote-exec": func() (ResourceProvisioner, error) {
			return new(MockResourceProvisioner), nil
		},
	}

	g, err := Graph(&GraphOpts{Module: m, Provisioners: provs})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get the resource we care about expanding
	n := g.Noun("aws_instance.web")
	if n == nil {
		t.Fatal("could not find")
	}
	rn := n.Meta.(*GraphNodeResource)

	g, err = rn.Expand()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTerraformGraphResourceExpandProvDepsStr)
	if actual != expected {
		t.Fatalf("bad:\n\nactual:\n%s\n\nexpected:\n%s", actual, expected)
	}
}

const testTerraformGraphStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> provider.aws
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
  aws_load_balancer.weblb -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
openstack_floating_ip.random
provider.aws
  provider.aws -> openstack_floating_ip.random
root
  root -> aws_instance.web
  root -> aws_load_balancer.weblb
  root -> aws_security_group.firewall
  root -> openstack_floating_ip.random
`

const testTerraformGraphCountStr = `
root: root
aws_instance.web
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
root
  root -> aws_instance.web
  root -> aws_load_balancer.weblb
`

const testTerraformGraphCountTaintedStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_instance.web.0 (tainted #1)
aws_instance.web.0 (tainted #1)
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
root
  root -> aws_instance.web
  root -> aws_instance.web.0 (tainted #1)
  root -> aws_load_balancer.weblb
`

const testTerraformGraphCountVarResourceStr = `
root: root
aws_instance.foo
aws_instance.web
  aws_instance.web -> aws_instance.foo
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
root
  root -> aws_instance.foo
  root -> aws_instance.web
  root -> aws_load_balancer.weblb
`

const testTerraformGraphDependsStr = `
root: root
aws_instance.db
  aws_instance.db -> aws_instance.web
aws_instance.web
root
  root -> aws_instance.db
  root -> aws_instance.web
`

const testTerraformGraphDependsCountStr = `
root: root
aws_instance.db
  aws_instance.db -> aws_instance.web
aws_instance.web
root
  root -> aws_instance.db
  root -> aws_instance.web
`

const testTerraformGraphDependsOrphanStr = `
root: root
aws_instance.db
  aws_instance.db -> aws_instance.web
aws_instance.old
aws_instance.web
root
  root -> aws_instance.db
  root -> aws_instance.old
  root -> aws_instance.web
`

const testTerraformGraphDiffStr = `
root: root
aws_instance.foo
root
  root -> aws_instance.foo
`

const testTerraformGraphDiffDestroyStr = `
root: root
aws_instance.bar
  aws_instance.bar -> aws_instance.bar (destroy)
  aws_instance.bar -> aws_instance.foo
  aws_instance.bar -> provider.aws
aws_instance.bar (destroy)
  aws_instance.bar (destroy) -> provider.aws
aws_instance.foo
  aws_instance.foo -> aws_instance.foo (destroy)
  aws_instance.foo -> provider.aws
aws_instance.foo (destroy)
  aws_instance.foo (destroy) -> aws_instance.bar (destroy)
  aws_instance.foo (destroy) -> provider.aws
provider.aws
root
  root -> aws_instance.bar
  root -> aws_instance.foo
`

const testTerraformGraphDiffDestroyCountsStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_instance.web (destroy)
aws_instance.web (destroy)
  aws_instance.web (destroy) -> aws_load_balancer.weblb (destroy)
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
  aws_load_balancer.weblb -> aws_load_balancer.weblb (destroy)
aws_load_balancer.weblb (destroy)
root
  root -> aws_instance.web
  root -> aws_load_balancer.weblb
`

const testTerraformGraphDiffModuleStr = `
root: root
aws_instance.foo
  aws_instance.foo -> aws_instance.foo (destroy)
  aws_instance.foo -> module.child
aws_instance.foo (destroy)
module.child
  module.child -> aws_instance.foo (destroy)
root
  root -> aws_instance.foo
  root -> module.child
`

const testTerraformGraphDiffModuleDependsStr = `
root: root
aws_instance.foo
  aws_instance.foo -> aws_instance.foo (destroy)
aws_instance.foo (destroy)
  aws_instance.foo (destroy) -> module.child
  aws_instance.foo (destroy) -> module.orphan
module.child
  module.child -> module.orphan
module.orphan
root
  root -> aws_instance.foo
  root -> module.child
  root -> module.orphan
`

const testTerraformGraphModulesStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> module.consul
  aws_instance.web -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
module.consul
  module.consul -> aws_security_group.firewall
  module.consul -> provider.aws
provider.aws
root
  root -> aws_instance.web
  root -> aws_security_group.firewall
  root -> module.consul
`

const testTerraformGraphModulesConsulStr = `
root: root
aws_instance.server
  aws_instance.server -> provider.aws
provider.aws
root
  root -> aws_instance.server
`

const testTerraformGraphModuleOrphanStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
module.consul
  module.consul -> provider.aws
provider.aws
root
  root -> aws_instance.web
  root -> aws_security_group.firewall
  root -> module.consul
`

const testTerraformGraphModuleOrphanConsulStr = `
root: root
aws_instance.old
  aws_instance.old -> provider.aws
provider.aws
root
  root -> aws_instance.old
`

const testTerraformGraphProviderPruneStr = `
root: root
aws_load_balancer.weblb
  aws_load_balancer.weblb -> provider.aws
provider.aws
root
  root -> aws_load_balancer.weblb`

const testTerraformGraphDiffCreateBeforeDestroyStr = `
root: root
aws_instance.bar
  aws_instance.bar -> provider.aws
aws_instance.bar (destroy)
  aws_instance.bar (destroy) -> aws_instance.bar
  aws_instance.bar (destroy) -> provider.aws
provider.aws
root
  root -> aws_instance.bar
  root -> aws_instance.bar (destroy)`

const testTerraformGraphStateStr = `
root: root
aws_instance.old
  aws_instance.old -> provider.aws
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> provider.aws
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
  aws_load_balancer.weblb -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
openstack_floating_ip.random
provider.aws
  provider.aws -> openstack_floating_ip.random
root
  root -> aws_instance.old
  root -> aws_instance.web
  root -> aws_load_balancer.weblb
  root -> aws_security_group.firewall
  root -> openstack_floating_ip.random
`

const testTerraformGraphTaintedStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_instance.web (tainted #1)
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> provider.aws
aws_instance.web (tainted #1)
  aws_instance.web (tainted #1) -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
provider.aws
root
  root -> aws_instance.web
  root -> aws_instance.web (tainted #1)
  root -> aws_security_group.firewall
`

const testTerraformGraphTaintedMultiStr = `
root: root
aws_instance.web
  aws_instance.web -> aws_instance.web (tainted #1)
  aws_instance.web -> aws_instance.web (tainted #2)
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> provider.aws
aws_instance.web (tainted #1)
  aws_instance.web (tainted #1) -> provider.aws
aws_instance.web (tainted #2)
  aws_instance.web (tainted #2) -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
provider.aws
root
  root -> aws_instance.web
  root -> aws_instance.web (tainted #1)
  root -> aws_instance.web (tainted #2)
  root -> aws_security_group.firewall
`

const testTerraformGraphCountOrphanStr = `
root: root
aws_instance.web
aws_load_balancer.old
  aws_load_balancer.old -> aws_instance.web
aws_load_balancer.weblb
  aws_load_balancer.weblb -> aws_instance.web
root
  root -> aws_instance.web
  root -> aws_load_balancer.old
  root -> aws_load_balancer.weblb
`

const testTerraformGraphOrphanModuleDepsStr = `
root: root
aws_instance.foo
  aws_instance.foo -> module.consul
  aws_instance.foo -> provider.aws
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> module.consul
  aws_instance.web -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
module.consul
  module.consul -> aws_security_group.firewall
  module.consul -> provider.aws
provider.aws
root
  root -> aws_instance.foo
  root -> aws_instance.web
  root -> aws_security_group.firewall
  root -> module.consul
`

const testTerraformGraphOrphanedModuleDepsStr = `
root: root
aws_instance.foo
  aws_instance.foo -> module.consul
  aws_instance.foo -> provider.aws
aws_instance.web
  aws_instance.web -> aws_security_group.firewall
  aws_instance.web -> module.consul
  aws_instance.web -> provider.aws
aws_security_group.firewall
  aws_security_group.firewall -> provider.aws
module.consul
  module.consul -> aws_security_group.firewall
  module.consul -> provider.aws
module.orphan
  module.orphan -> aws_instance.foo
  module.orphan -> aws_instance.web
  module.orphan -> provider.aws
provider.aws
root
  root -> aws_instance.foo
  root -> aws_instance.web
  root -> aws_security_group.firewall
  root -> module.consul
  root -> module.orphan
`

const testTerraformGraphResourceExpandStr = `
root: root
aws_instance.web.0
aws_instance.web.1
aws_instance.web.2
root
  root -> aws_instance.web.0
  root -> aws_instance.web.1
  root -> aws_instance.web.2
`

const testTerraformGraphResourceExpandProvDepsStr = `
root: root
aws_instance.web.0
aws_instance.web.1
  aws_instance.web.1 -> aws_instance.web.0
aws_instance.web.2
  aws_instance.web.2 -> aws_instance.web.0
root
  root -> aws_instance.web.0
  root -> aws_instance.web.1
  root -> aws_instance.web.2
`
