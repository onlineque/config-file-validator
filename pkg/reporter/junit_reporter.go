package reporter

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type JunitReporter struct{}

const (
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

// https://github.com/testmoapp/junitxml#basic-junit-xml-structure
type Testsuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	Name       string      `xml:"name,attr,omitempty"`
	Tests      int         `xml:"tests,attr,omitempty"`
	Failures   int         `xml:"failures,attr,omitempty"`
	Errors     int         `xml:"errors,attr,omitempty"`
	Skipped    int         `xml:"skipped,attr,omitempty"`
	Assertions int         `xml:"assertions,attr,omitempty"`
	Time       float32     `xml:"time,attr,omitempty"`
	Timestamp  *time.Time  `xml:"timestamp,attr,omitempty"`
	Testsuites []Testsuite `xml:"testsuite"`
}

type Testsuite struct {
	XMLName    xml.Name    `xml:"testsuite"`
	Name       string      `xml:"name,attr"`
	Tests      int         `xml:"tests,attr,omitempty"`
	Failures   int         `xml:"failures,attr,omitempty"`
	Errors     int         `xml:"errors,attr,omitempty"`
	Skipped    int         `xml:"skipped,attr,omitempty"`
	Assertions int         `xml:"assertions,attr,omitempty"`
	Time       float32     `xml:"time,attr,omitempty"`
	Timestamp  *time.Time  `xml:"timestamp,attr,omitempty"`
	File       string      `xml:"file,attr,omitempty"`
	Testcases  *[]Testcase `xml:"testcase,omitempty"`
	Properties *[]Property `xml:"properties>property,omitempty"`
	SystemOut  *SystemOut  `xml:"system-out,omitempty"`
	SystemErr  *SystemErr  `xml:"system-err,omitempty"`
}

type Testcase struct {
	XMLName         xml.Name         `xml:"testcase"`
	Name            string           `xml:"name,attr"`
	ClassName       string           `xml:"classname,attr"`
	Assertions      int              `xml:"assertions,attr,omitempty"`
	Time            float32          `xml:"time,attr,omitempty"`
	File            string           `xml:"file,attr,omitempty"`
	Line            int              `xml:"line,attr,omitempty"`
	Skipped         *Skipped         `xml:"skipped,omitempty,omitempty"`
	Properties      *[]Property      `xml:"properties>property,omitempty"`
	TestcaseError   *TestcaseError   `xml:"error,omitempty"`
	TestcaseFailure *TestcaseFailure `xml:"failure,omitempty"`
}

type Skipped struct {
	XMLName xml.Name `xml:"skipped"`
	Message string   `xml:"message,attr"`
}

type TestcaseError struct {
	XMLName   xml.Name `xml:"error"`
	Message   string   `xml:"message,omitempty"`
	Type      string   `xml:"type,omitempty"`
	TextValue string   `xml:",chardata"`
}

type TestcaseFailure struct {
	XMLName   xml.Name `xml:"failure"`
	Message   string   `xml:"message,omitempty"`
	Type      string   `xml:"type,omitempty"`
	TextValue string   `xml:",chardata"`
}

type SystemOut struct {
	XMLName   xml.Name `xml:"system-out"`
	TextValue string   `xml:",chardata"`
}

type SystemErr struct {
	XMLName   xml.Name `xml:"system-err"`
	TextValue string   `xml:",chardata"`
}

type Property struct {
	XMLName   xml.Name `xml:"property"`
	TextValue string   `xml:",chardata"`
	Name      string   `xml:"name,attr"`
	Value     string   `xml:"value,attr,omitempty"`
}

func (ts Testsuites) checkPropertyValidity() error {
	for tsidx := range ts.Testsuites {
		testsuite := ts.Testsuites[tsidx]
		if testsuite.Properties == nil {
			continue
		}
		for pridx := range *testsuite.Properties {
			property := (*testsuite.Properties)[pridx]
			if property.Value != "" && property.TextValue != "" {
				return fmt.Errorf("property %s in testsuite %s should contain value or a text value, not both",
					property.Name, testsuite.Name)
			}
		}
		if testsuite.Testcases == nil {
			continue
		}
		for tcidx := range *testsuite.Testcases {
			testcase := (*testsuite.Testcases)[tcidx]
			if testcase.Properties == nil {
				continue
			}
			for propidx := range *testcase.Properties {
				property := (*testcase.Properties)[propidx]
				if property.Value != "" && property.TextValue != "" {
					return fmt.Errorf("property %s in testcase %s should contain value or a text value, not both",
						property.Name, testcase.Name)
				}
			}
		}
	}
	return nil
}

func (ts Testsuites) getReport() ([]byte, error) {
	err := ts.checkPropertyValidity()
	if err != nil {
		return []byte{}, err
	}

	data, err := xml.MarshalIndent(ts, " ", "  ")
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (jr JunitReporter) Print(reports []Report) error {
	testcases := []Testcase{}
	testErrors := 0

	for _, r := range reports {
		if strings.Contains(r.FilePath, "\\") {
			r.FilePath = strings.ReplaceAll(r.FilePath, "\\", "/")
		}
		tc := Testcase{Name: fmt.Sprintf("%s validation", r.FilePath), File: r.FilePath, ClassName: "config-file-validator"}
		if !r.IsValid {
			testErrors++
			tc.TestcaseFailure = &TestcaseFailure{Message: r.ValidationError.Error()}
		}

		testcases = append(testcases, tc)
		// fmt.Println(r.FilePath, r.FileName, r.IsValid, r.ValidationError)
	}
	testsuite := Testsuite{Name: "config-file-validator", Testcases: &testcases, Errors: testErrors}
	testsuiteBatch := []Testsuite{testsuite}
	ts := Testsuites{Name: "config-file-validator", Tests: len(reports), Testsuites: testsuiteBatch}

	data, err := ts.getReport()
	if err != nil {
		return err
	}
	fmt.Println(Header + string(data))
	return nil
}
