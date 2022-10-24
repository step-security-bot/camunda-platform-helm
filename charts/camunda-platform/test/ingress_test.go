package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
)

type ingressTemplateTest struct {
	suite.Suite
	chartPath string
	release   string
	namespace string
	templates []string
}

func TestIngressTemplate(t *testing.T) {
	t.Parallel()

	chartPath, err := filepath.Abs("../")
	require.NoError(t, err)

	suite.Run(t, &ingressTemplateTest{
		chartPath: chartPath,
		release:   "camunda-platform-test",
		namespace: "camunda-platform-" + strings.ToLower(random.UniqueId()),
		templates: []string{"templates/ingress.yaml"},
	})
}

func (s *ingressTemplateTest) TestIngressWithKeycloakChart() {
	// given
	options := &helm.Options{
		SetValues: map[string]string{
			"global.ingress.enabled":    "true",
			"identity.keycloak.enabled": "true",
			"identity.contextPath":      "/identity",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
		ExtraArgs:      map[string][]string{"template": {"--debug"}, "install": {"--debug"}},
	}

	// NOTE: helm.Options.ExtraArgs doesn't support passing args to Helm "template" command.
	// TODO: Remove "template" from all helm.Options.ExtraArgs since it doesn't have any effect.
	extraArgs := []string{"--show-only", "charts/identity/charts/keycloak/templates/statefulset.yaml"}

	// when
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, nil, extraArgs...)

	var statefulSet v1.StatefulSet
	helm.UnmarshalK8SYaml(s.T(), output, &statefulSet)

	// then
	env := statefulSet.Spec.Template.Spec.Containers[0].Env
	s.Require().Contains(env,
		v12.EnvVar{
			Name:  "KEYCLOAK_PROXY_ADDRESS_FORWARDING",
			Value: "true",
		})
}

func (s *ingressTemplateTest) TestIngressWithContextPath() {
	// given
	options := &helm.Options{
		SetValues: map[string]string{
			"global.ingress.enabled": "true",
			"identity.contextPath":   "/identity",
			"operate.contextPath":    "/operate",
			"optimize.contextPath":   "/optimize",
			"tasklist.contextPath":   "/tasklist",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
		ExtraArgs:      map[string][]string{"template": {"--debug"}, "install": {"--debug"}},
	}

	// when
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)

	// then
	s.Require().Contains(output, "kind: Ingress")
	s.Require().Contains(output, "path: /auth")
	s.Require().Contains(output, "path: /identity")
	s.Require().Contains(output, "path: /operate")
	s.Require().Contains(output, "path: /optimize")
	s.Require().Contains(output, "path: /tasklist")
}

func (s *ingressTemplateTest) TestIngressComponentWithNoContextPath() {
	// given
	options := &helm.Options{
		SetValues: map[string]string{
			"global.ingress.enabled": "true",
			"identity.contextPath":   "",
			"operate.contextPath":    "",
			"optimize.contextPath":   "",
			"tasklist.contextPath":   "",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
		ExtraArgs:      map[string][]string{"template": {"--debug"}, "install": {"--debug"}},
	}

	// when
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)

	// then
	s.Require().NotContains(output, "name: camunda-platform-test-identity")
	s.Require().NotContains(output, "name: camunda-platform-test-operate")
	s.Require().NotContains(output, "name: camunda-platform-test-optimize")
	s.Require().NotContains(output, "name: camunda-platform-test-tasklist")
}

func (s *ingressTemplateTest) TestIngressComponentDisabled() {
	// given
	options := &helm.Options{
		SetValues: map[string]string{
			"global.ingress.enabled": "true",
			"operate.identity":       "false",
			"operate.enabled":        "false",
			"optimize.enabled":       "false",
			"tasklist.enabled":       "false",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", s.namespace),
		ExtraArgs:      map[string][]string{"template": {"--debug"}, "install": {"--debug"}},
	}

	// when
	output := helm.RenderTemplate(s.T(), options, s.chartPath, s.release, s.templates)

	// then
	s.Require().NotContains(output, "name: camunda-platform-test-identity")
	s.Require().NotContains(output, "name: camunda-platform-test-operate")
	s.Require().NotContains(output, "name: camunda-platform-test-optimize")
	s.Require().NotContains(output, "name: camunda-platform-test-tasklist")
}