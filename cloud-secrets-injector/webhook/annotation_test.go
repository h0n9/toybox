package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	SampleTemplate = `
{{ range $k, $v := . }}export {{ $k }}={{ $v }}
{{ end }}
	`
)

func TestParseAndCheckAnnotations(t *testing.T) {
	input := map[string]string{
		"cloud-secrets-injector.h0n9.postie.chat/provider":          "aws",               // ✅
		"cloud-secrets-injector.h0n9.postie.chat/key-id":            "life-is-beautiful", // ✅
		"cloud-secrets-injector.h0n9.postie.chat/output":            "/envs",             // ✅
		"cloud-secrets-injector.h0n9.postie.chat/template":          SampleTemplate,      // ✅
		"cloud-secrets-injector.h0n9.postie.chat/injected":          "true",              // ✅
		"cloud-secrets-injector.h0n9.posite.chat/template":          SampleTemplate,      // ❌: typo
		"cloud-secrets-injector.h0n9.postie.chat/volume-path":       "/envs",             // ❌: unsupported
		"cloud-secrets-injector.h0n9.postie.chat":                   "h0n9",              // ❌: non subpath
		"vault.hashicorp.com/secret-volume-path-SECRET-NAME-foobar": "/envs",             // ❌: non related annotation
	}
	output := ParseAndCheckAnnotations(input)
	expectedOutput := map[string]string{
		"provider": "aws",
		"key-id":   "life-is-beautiful",
		"template": SampleTemplate,
		"output":   "/envs",
		"injected": "true",
	}
	assert.EqualValues(t, expectedOutput, output)
}
