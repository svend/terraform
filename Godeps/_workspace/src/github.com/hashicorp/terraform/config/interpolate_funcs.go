package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// Funcs is the mapping of built-in functions for configuration.
var Funcs map[string]InterpolationFunc

func init() {
	Funcs = map[string]InterpolationFunc{
		"concat": interpolationFuncConcat,
		"file":   interpolationFuncFile,
		"join":   interpolationFuncJoin,
		"lookup": interpolationFuncLookup,
		"element":  interpolationFuncElement,
	}
}

// interpolationFuncConcat implements the "concat" function that allows
// strings to be joined together.
func interpolationFuncConcat(
	vs map[string]string, args ...string) (string, error) {
	var buf bytes.Buffer

	for _, a := range args {
		if _, err := buf.WriteString(a); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// interpolationFuncFile implements the "file" function that allows
// loading contents from a file.
func interpolationFuncFile(
	vs map[string]string, args ...string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf(
			"file expects 1 arguments, got %d", len(args))
	}

	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// interpolationFuncJoin implements the "join" function that allows
// multi-variable values to be joined by some character.
func interpolationFuncJoin(
	vs map[string]string, args ...string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("join expects 2 arguments")
	}

	var list []string
	for _, arg := range args[1:] {
		parts := strings.Split(arg, InterpSplitDelim)
		list = append(list, parts...)
	}

	return strings.Join(list, args[0]), nil
}

// interpolationFuncLookup implements the "lookup" function that allows
// dynamic lookups of map types within a Terraform configuration.
func interpolationFuncLookup(
	vs map[string]string, args ...string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf(
			"lookup expects 2 arguments, got %d", len(args))
	}

	k := fmt.Sprintf("var.%s", strings.Join(args, "."))
	v, ok := vs[k]
	if !ok {
		return "", fmt.Errorf(
			"lookup in '%s' failed to find '%s'",
			args[0], args[1])
	}

	return v, nil
}

// interpolationFuncElement implements the "element" function that allows
// a specific index to be looked up in a multi-variable value. Note that this will
// wrap if the index is larger than the number of elements in the multi-variable value.
func interpolationFuncElement(
	vs map[string]string, args ...string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf(
			"element expects 2 arguments, got %d", len(args))
	}

	list := strings.Split(args[0], InterpSplitDelim)

	index, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf(
			"invalid number for index, got %s", args[1])
	}

	v := list[index % len(list)]

	return v, nil
}
