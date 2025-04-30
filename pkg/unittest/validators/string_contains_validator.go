package validators

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"gopkg.in/yaml.v3"
)

// StringContainsValidator validates whether the value at Path is a string and contains the specified Content
type StringContainsValidator struct {
	Path             string
	Content          interface{}
	IgnoreFormatting bool // When true, ignores spaces, tabs and line breaks in comparison
	FromJson         bool // When true, treats the string as JSON and checks if it contains the YAML content
	FromYaml         bool // When true, treats the string as YAML and checks if it contains the YAML content
}

// normalizeWhitespace removes all whitespace characters and replaces them with a single space
func normalizeWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

func (v StringContainsValidator) failInfo(actual interface{}, manifestIndex, assertIndex int, not bool) []string {
	var expectedStr string
	if str, ok := v.Content.(string); ok {
		expectedStr = str
	} else {
		expectedStr = common.TrustedMarshalYAML(v.Content)
	}

	var actualStr string
	if str, ok := actual.(string); ok {
		actualStr = str
	} else {
		actualStr = common.TrustedMarshalYAML(actual)
	}
	containsFailFormat := setFailFormat(not, true, true, false, " to contain")

	log.WithField("validator", "stringContains").Debugln("expected content:", expectedStr)
	log.WithField("validator", "stringContains").Debugln("actual string:", actualStr)
	log.WithField("validator", "stringContains").Debugln("ignoreFormatting:", v.IgnoreFormatting)
	log.WithField("validator", "stringContains").Debugln("fromJson:", v.FromJson)
	log.WithField("validator", "stringContains").Debugln("fromYaml:", v.FromYaml)

	return splitInfof(
		containsFailFormat,
		manifestIndex,
		assertIndex,
		v.Path,
		expectedStr,
		actualStr,
	)
}

// containsMap checks if bigMap contains all keys and values in smallMap
func containsMap(bigMap, smallMap map[string]interface{}) bool {
	for k, smallVal := range smallMap {
		bigVal, exists := bigMap[k]
		if !exists {
			return false
		}

		// Handle nested maps
		if smallValMap, smallIsMap := smallVal.(map[string]interface{}); smallIsMap {
			if bigValMap, bigIsMap := bigVal.(map[string]interface{}); bigIsMap {
				if !containsMap(bigValMap, smallValMap) {
					return false
				}
			} else {
				return false
			}
		} else if !reflect.DeepEqual(bigVal, smallVal) {
			return false
		}
	}

	return true
}

func (v StringContainsValidator) validateJSON(jsonStr string, manifestIndex, assertIndex int, context *ValidateContext) (bool, []string) {
	validateErrors := []string{}

	// Parse JSON string into a map
	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
			fmt.Sprintf("failed to parse JSON from '%s': %s", v.Path, err.Error()))
		return false, validateErrors
	}

	// Convert content to a map
	var contentMap map[string]interface{}

	switch content := v.Content.(type) {
	case map[string]interface{}:
		contentMap = content
	case string:
		if err := yaml.Unmarshal([]byte(content), &contentMap); err != nil {
			validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
				fmt.Sprintf("failed to parse Content as YAML: %s", err.Error()))
			return false, validateErrors
		}
	default:
		// Convert arbitrary content to YAML and then to map
		contentYAML := common.TrustedMarshalYAML(v.Content)
		if err := yaml.Unmarshal([]byte(contentYAML), &contentMap); err != nil {
			validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
				fmt.Sprintf("failed to convert Content to map: %s", err.Error()))
			return false, validateErrors
		}
	}

	// Check if JSON contains the content map
	found := containsMap(jsonObj, contentMap)

	if found == context.Negative {
		validateErrors = v.failInfo(jsonStr, manifestIndex, assertIndex, context.Negative)
		return false, validateErrors
	}

	return true, validateErrors
}

func (v StringContainsValidator) validateYAML(yamlStr string, manifestIndex, assertIndex int, context *ValidateContext) (bool, []string) {
	validateErrors := []string{}

	// Parse YAML string into a map
	var yamlObj map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &yamlObj); err != nil {
		validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
			fmt.Sprintf("failed to parse YAML from '%s': %s", v.Path, err.Error()))
		return false, validateErrors
	}

	// Convert content to a map
	var contentMap map[string]interface{}

	switch content := v.Content.(type) {
	case map[string]interface{}:
		contentMap = content
	case string:
		if err := yaml.Unmarshal([]byte(content), &contentMap); err != nil {
			validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
				fmt.Sprintf("failed to parse Content as YAML: %s", err.Error()))
			return false, validateErrors
		}
	default:
		// Convert arbitrary content to YAML and then to map
		contentYAML := common.TrustedMarshalYAML(v.Content)
		if err := yaml.Unmarshal([]byte(contentYAML), &contentMap); err != nil {
			validateErrors = splitInfof(errorFormat, manifestIndex, assertIndex,
				fmt.Sprintf("failed to convert Content to map: %s", err.Error()))
			return false, validateErrors
		}
	}

	// Check if YAML contains the content map
	found := containsMap(yamlObj, contentMap)

	if found == context.Negative {
		validateErrors = v.failInfo(yamlStr, manifestIndex, assertIndex, context.Negative)
		return false, validateErrors
	}

	return true, validateErrors
}

func (v StringContainsValidator) validateSingle(singleActual string, manifestIndex, assertIndex int, context *ValidateContext) (bool, []string) {
	// If we're treating the string as JSON or YAML
	if v.FromJson {
		return v.validateJSON(singleActual, manifestIndex, assertIndex, context)
	}

	if v.FromYaml {
		return v.validateYAML(singleActual, manifestIndex, assertIndex, context)
	}

	validateSingleErrors := []string{}

	var found bool
	// If content is not string, convert it to string
	contentStr, isString := v.Content.(string)
	if !isString {
		contentStr = common.TrustedMarshalYAML(v.Content)
	}

	if v.IgnoreFormatting {
		// Normalize whitespace in both strings before comparing
		normalizedActual := normalizeWhitespace(singleActual)
		normalizedContent := normalizeWhitespace(contentStr)
		found = strings.Contains(normalizedActual, normalizedContent)
	} else {
		// Standard comparison
		found = strings.Contains(singleActual, contentStr)
	}

	if found == context.Negative {
		validateSingleErrors = v.failInfo(singleActual, manifestIndex, assertIndex, context.Negative)
		return false, validateSingleErrors
	}

	return true, validateSingleErrors
}

func (v StringContainsValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	manifestSuccess := (len(actual) == 0 && context.Negative)
	var manifestValidateErrors []string

	for valuesIndex, singleActual := range actual {
		singleSuccess := false
		var singleValidateErrors []string

		// Convert to string if possible
		switch actualValue := singleActual.(type) {
		case string:
			singleSuccess, singleValidateErrors = v.validateSingle(actualValue, manifestIndex, valuesIndex, context)
		default:
			// Try to convert to string using YAML marshalling for non-string values
			actualStr := common.TrustedMarshalYAML(singleActual)
			singleSuccess, singleValidateErrors = v.validateSingle(actualStr, manifestIndex, valuesIndex, context)
		}

		manifestValidateErrors = append(manifestValidateErrors, singleValidateErrors...)
		manifestSuccess = determineSuccess(valuesIndex, manifestSuccess, singleSuccess)

		if !manifestSuccess && context.FailFast {
			break
		}
	}

	return manifestSuccess, manifestValidateErrors
}

// Validate implements Validatable
func (v StringContainsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		manifestSuccess, manifestValidateErrors := v.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, manifestValidateErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, manifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo("no manifest found", 0, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
