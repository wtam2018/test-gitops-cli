package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mkmik/multierror"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/ioutils"
	"knative.dev/pkg/apis"
)

const (
	DNS1035Error = "a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
)

func TestValidate(t *testing.T) {

	tests := []struct {
		desc     string
		filename string
		wantErr  error
	}{
		{
			"service repo URL must be the same Git type as the GitOps URL",
			"testdata/svc_git_type_mismatch.yaml",
			multierror.Join(
				[]error{
					inconsistentGitTypeError("github", "https://gitlab.com/myproject/myservice.git", []string{"environments.test-dev.apps.bus.services.bus-svc"}),
				},
			),
		},
		{
			"Environment Duplicate Name entry",
			"testdata/environment_config_name.yaml",
			multierror.Join(
				[]error{
					invalidEnvironment("argocd", "Environment name cannot be the same as a config name.", []string{"environments.argocd"}),
					invalidEnvironment("tst-cicd", "Environment name cannot be the same as a config name.", []string{"environments.tst-cicd"}),
				},
			),
		},
		{
			"Invalid entity name error",
			"testdata/name_error.yaml",
			multierror.Join(
				[]error{
					invalidNameError("argo.cd", DNS1035Error, []string{"config.argocd"}),
					invalidNameError("tst!cicd", DNS1035Error, []string{"config.tst!cicd"}),
					invalidNameError("", DNS1035Error, []string{"environments.develo.pment.apps.app-1$.services"}),
					invalidNameError("", DNS1035Error, []string{"environments.develo.pment.apps.app-1$.services.pipelines.integration.binding"}),
					invalidNameError("app-1$", DNS1035Error, []string{"environments.develo.pment.apps.app-1$"}),
					invalidNameError("develo.pment", DNS1035Error, []string{"environments.develo.pment"}),
				},
			),
		},
		{
			"Invalid long service name error",
			"testdata/service_name_long.yaml",
			multierror.Join(
				[]error{
					invalidNameError("my-incredibly-long-name-for-a-test-service-that-fails", LongServiceNameError, []string{"environments.development.apps.app-1.services.my-incredibly-long-name-for-a-test-service-that-fails"}),
				},
			),
		},
		{
			"Missing field error",
			"testdata/missing_fields_error.yaml",
			multierror.Join([]error{
				missingFieldsError([]string{"secret"}, []string{"environments.development.apps.app-1.services.service-1.webhook"}),
				missingFieldsError([]string{"integration"}, []string{"environments.development.apps.app-1.services.service-1.pipelines"}),
			}),
		},
		{
			"Missing service and config repo from application",
			"testdata/missing_service_error.yaml",
			multierror.Join([]error{
				missingFieldsError([]string{"services", "config_repo"}, []string{"environments.development.apps.app-1"}), missingFieldsError([]string{"path"}, []string{"environments.development.apps.app-2.config_repo"}),
				missingFieldsError([]string{"url"}, []string{"environments.development.apps.app-3.config_repo"}),
				missingFieldsError([]string{"url", "path"}, []string{"environments.development.apps.app-4.config_repo"}),
				apis.ErrMultipleOneOf("environments.development.apps.app-5.services", "environments.development.apps.app-5.config_repo"),
			}),
		},
		{
			"duplicate environment name error",
			"testdata/duplicate_environment.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"duplicate-environment"}, []string{"environments.duplicate-environment"}),
				},
			),
		},
		{
			"duplicate application name error",
			"testdata/duplicate_application.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"my-app-1"}, []string{"environments.app-environment.apps.my-app-1"}),
				},
			),
		},
		{
			"duplicate service name error",
			"testdata/duplicate_service.yaml",
			multierror.Join(
				[]error{
					duplicateFieldsError([]string{"app-1-service-http"}, []string{"environments.duplicate-service.apps.my-app-2.services.app-1-service-http"}),
				},
			),
		},
		{
			"missing app service reference",
			"testdata/duplicate_source_url.yaml",
			multierror.Join(
				[]error{
					duplicateSourceError("https://github.com/testing/testing.git", []string{"environments.duplicate-source.apps.my-app-1.services.app-1-service-http", "environments.duplicate-source.apps.my-app-2.services.app-2-service-http"}),
				},
			),
		},
		{
			"service with pipeline with no template",
			"testdata/service_with_bindings_no_template.yaml",
			nil,
		},
		{
			"valid manifest file",
			"testdata/valid_manifest.yaml",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s (%s)", test.desc, test.filename), func(rt *testing.T) {
			pipelines, err := ParseFile(ioutils.NewFilesystem(), test.filename)
			if err != nil {
				rt.Fatalf("failed to parse file:%v", err)
			}
			got := pipelines.Validate()
			err = matchMultiErrors(rt, got, test.wantErr)
			if err != nil {
				rt.Fatal(err)
			}
		})
	}
}

func matchMultiErrors(t *testing.T, a error, b error) error {
	t.Helper()
	if a == nil || b == nil {
		if a != b {
			return fmt.Errorf("did not match error, got %v want %v", a, b)
		}
		return nil
	}
	got, want := multierror.Split(a), multierror.Split(b)
	if len(got) != len(want) {
		return fmt.Errorf("error count did not match, got %d want %d", len(got), len(want))
	}
	for i := 0; i < len(got); i++ {
		if diff := cmp.Diff(got[i].Error(), want[i].Error()); diff != "" {
			return fmt.Errorf("did not match error, got %v want %v", got[i], want[i])
		}
	}
	return nil
}
