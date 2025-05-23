package results

import (
	"fmt"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
)

// AssertionResult result return by Assertion.Assert
type AssertionResult struct {
	Index      int
	FailInfo   []string
	Passed     bool
	Skipped    bool
	SkipReason string
	AssertType string
	Not        bool
	CustomInfo string
}

func (ar AssertionResult) print(printer *printer.Printer, verbosity int) {
	if ar.Passed {
		return
	}

	printer.Println(printer.Danger("%s", ar.getTitle()), 2)
	for _, infoLine := range ar.FailInfo {
		printer.Println(infoLine, 3)
	}
	printer.Println("", 0)
}

func (ar AssertionResult) getTitle() string {
	var title string
	if ar.CustomInfo != "" {
		title = ar.CustomInfo
	} else {
		var notAnnotation string
		if ar.Not {
			notAnnotation = " NOT"
		}
		title = fmt.Sprintf("- asserts[%d]%s `%s` fail", ar.Index, notAnnotation, ar.AssertType)
	}
	return title
}

// ToString writing the object to a customized formatted string.
func (ar AssertionResult) stringify() string {
	content := fmt.Sprintf("\t\t %s \n", ar.getTitle())

	for _, infoLine := range ar.FailInfo {
		content += fmt.Sprintf("\t\t\t %s \n", infoLine)
	}

	return content
}
