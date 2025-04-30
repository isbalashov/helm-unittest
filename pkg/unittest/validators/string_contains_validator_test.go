package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"

	"github.com/stretchr/testify/assert"
)

var docToTestStringContains = `
a:
  b: "hello world foo bar"
  c: "multi\nline\nstring"
  d: '{"name":"test","nested":{"value":true}}'
  e: |
    some:
      nested: yaml
      format: true
`

func TestStringContainsValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorWhenEmptyManifestFail(t *testing.T) {
	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0",
		"Path:\ta.b",
		"Expected to contain:",
		"\thello world",
		"Actual:", "\tno manifest found"}, diff)
}

func TestStringContainsValidatorWhenEmptyManifestNegativeOk(t *testing.T) {
	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "not present",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "not present",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	not present",
		"Actual:",
		"	hello world foo bar",
	}, diff)
}

func TestStringContainsValidatorWithIgnoreFormattingWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:             "a.c",
		Content:          "multi line string",
		IgnoreFormatting: true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
func TestStringContainsValidatorWithIgnoreFormattingWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:             "a.c",
		Content:          "not present",
		IgnoreFormatting: true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.c",
		"Expected to contain:",
		"	not present",
		"Actual:",
		"	multi",
		"	line",
		"	string",
	}, diff)
}

func TestStringContainsValidatorFromJsonWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.d",
		FromJson: true,
		Content: map[string]interface{}{
			"name": "test",
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorFromJsonNestedWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.d",
		FromJson: true,
		Content: map[string]interface{}{
			"nested": map[string]interface{}{
				"value": true,
			},
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorFromJsonWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.d",
		FromJson: true,
		Content: map[string]interface{}{
			"notfound": "value",
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.d",
		"Expected to contain:",
		"	notfound: value",
		"Actual:",
		"	{\"name\":\"test\",\"nested\":{\"value\":true}}",
	}, diff)
}

func TestStringContainsValidatorFromJsonInvalidJson(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.b",
		FromJson: true,
		Content: map[string]interface{}{
			"key": "value",
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Contains(t, diff, "Error:")

	errorFound := false
	for i, line := range diff {
		if line == "Error:" && i+1 < len(diff) {
			assert.Contains(t, diff[i+1], "failed to parse JSON from")
			errorFound = true
			break
		}
	}
	assert.True(t, errorFound, "Error message about JSON parsing not found in the diff output")
}
func TestStringContainsValidatorFromYamlWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.e",
		FromYaml: true,
		Content: map[string]interface{}{
			"some": map[string]interface{}{
				"nested": "yaml",
			},
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorFromYamlWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:     "a.e",
		FromYaml: true,
		Content: map[string]interface{}{
			"notfound": "value",
		},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.e",
		"Expected to contain:",
		"	notfound: value",
		"Actual:",
		"	some:",
		"	  nested: yaml",
		"	  format: true",
	}, diff)
}

func TestStringContainsValidatorMultiManifestWhenOk(t *testing.T) {
	manifest1 := makeManifest(docToTestStringContains)
	manifest2 := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestStringContains)
	extraDoc := `
a:
  b: "different string"
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: manifests,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	hello world",
		"Actual:",
		"	different string",
	}, diff)
}

func TestStringContainsValidatorMultiManifestWhenFailFast(t *testing.T) {
	manifest1 := makeManifest(docToTestStringContains)
	extraDoc := `
a:
  b: "different string"
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest2, manifest1}

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     manifests,
		FailFast: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	hello world",
		"Actual:",
		"	different string",
	}, diff)
}

func TestStringContainsValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected NOT to contain:",
		"	hello world",
		"Actual:",
		"	hello world foo bar",
	}, diff)
}

func TestStringContainsValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.nonexistent",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.nonexistent",
	}, diff)
}

func TestStringContainsValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	validator := StringContainsValidator{
		Path:    "a.nonexistent",
		Content: "hello world",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestStringContainsValidatorContentAsStructuredData(t *testing.T) {
	manifest := makeManifest(docToTestStringContains)

	content := map[string]interface{}{
		"key": "value",
	}

	validator := StringContainsValidator{
		Path:    "a.b",
		Content: content,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Contains(t, diff[4], "key: value")
}
