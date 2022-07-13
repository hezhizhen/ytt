// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validations_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/vmware-tanzu/carvel-ytt/pkg/experiments"
	"github.com/vmware-tanzu/carvel-ytt/pkg/validations"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
	_ "github.com/vmware-tanzu/carvel-ytt/pkg/yttlibraryext"
	"github.com/vmware-tanzu/carvel-ytt/test/filetests"
)

var (
	// Example usage:
	//   Run a specific test:
	//   ./hack/test-all.sh -v -run TestYAMLTemplate/filetests/if.tpltest
	//
	//   Include template compilation results in the output:
	//   ./hack/test-all.sh -v -run TestYAMLTemplate/filetests/if.tpltest TestYAMLTemplate.code=true
	showTemplateCodeFlag = kvArg("TestYAMLTemplate.code")
)

// TestMain is invoked when any tests are run in this package, *instead of* those tests being run directly.
// This allows for setup to occur before *any* test is run.
func TestMain(m *testing.M) {
	experiments.ResetForTesting()
	os.Setenv(experiments.Env, "validations")

	exitVal := m.Run() // execute the specified tests

	os.Exit(exitVal) // required in order to properly report the error level when tests fail.
}

func TestYAMLTemplate(t *testing.T) {
	ft := filetests.FileTests{}
	ft.PathToTests = "filetests"
	ft.ShowTemplateCode = showTemplateCode(kvArg("TestYAMLTemplate.code"))
	ft.EvalFunc = EvalAndValidateTemplate(ft)

	ft.Run(t)
}

func EvalAndValidateTemplate(ft filetests.FileTests) filetests.EvaluateTemplate {
	return func(src string) (filetests.MarshalableResult, *filetests.TestErr) {
		result, testErr := ft.DefaultEvalTemplate(src)

		err := validations.ProcessAssertValidateAnns(result.(yamlmeta.Node))
		if err != nil {
			return nil, filetests.NewTestErr(err, fmt.Errorf("Failed to process @assert/validate annotations."))
		}

		chk := validations.Run(result.(yamlmeta.Node), "template-test")
		if chk.HasViolations() {
			err := fmt.Errorf("\n%s", chk.Error())
			return nil, filetests.NewTestErr(err, fmt.Errorf("validation error: %v\n", err))
		}

		return result, testErr
	}
}

func showTemplateCode(showTemplateCodeFlag string) bool {
	return strings.HasPrefix(strings.ToLower(showTemplateCodeFlag), "t")
}

func kvArg(name string) string {
	name += "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, name) {
			return strings.TrimPrefix(arg, name)
		}
	}
	return ""
}
