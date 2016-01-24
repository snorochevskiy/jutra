package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type TestSuite struct {
	Name       string         `xml:"name,attr"`
	Properties *PropertiesTag `xml:"properties"`
	TestCases  []TestCase     `xml:"testcase"`

	TestsNumber  string `xml:"tests,attr"`
	TestsSkipped string `xml:"skipped,attr"`
	TestsFailed  string `xml:"failures,attr"`
	TestsErrors  string `xml:"errors,attr"`

	// Error of whole suite
	SystemErr string `xml:"system-err,omitempty"`
}

type PropertiesTag struct {
}

type TestCase struct {
	Name      string         `xml:"name,attr"`
	ClassName string         `xml:"classname,attr"`
	Skipped   *SkippedStatus `xml:"skipped"`
	Failure   *FailureStatus `xml:"failure"`

	Md5Hash        string
	TestCaseStatus string
}

type SkippedStatus struct {
}

type FailureStatus struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

func (testcase *TestCase) IsSkipped() bool {
	return testcase.Skipped != nil
}

func ProcessAllResultsFiles(branch string, fullDirPath string) {
	reportFiles, reportFilesErr := ioutil.ReadDir(fullDirPath)
	if reportFilesErr != nil {
		log.Panic(reportFilesErr)
		return
	}

	allTests := make([]*TestCase, 0, 100)

	for _, reportFile := range reportFiles {
		if !reportFile.IsDir() && strings.HasSuffix(reportFile.Name(), ".xml") {
			fullReportFilePath := path.Join(fullDirPath, reportFile.Name())
			suite, suitePathErr := ParseTestSuite(fullReportFilePath)
			if suitePathErr != nil {
				log.Println(suitePathErr)
				continue
			}

			for i := 0; i < len(suite.TestCases); i++ {

				test := suite.TestCases[i]
				if test.Failure != nil {
					test.TestCaseStatus = TEST_CASE_STATUS_FAILED
				} else if test.Skipped != nil {
					test.TestCaseStatus = TEST_CASE_STATUS_SKIPPED
				} else {
					test.TestCaseStatus = TEST_CASE_STATUS_PASSED
				}

				md5Hash := md5.Sum([]byte(test.ClassName + "#" + test.Name))
				test.Md5Hash = hex.EncodeToString(md5Hash[:])
				allTests = append(allTests, &test)
			}
		}
	}

	DAO.PersistLaunch(branch, allTests)
}

func ParseTestSuite(fullFilePath string) (*TestSuite, error) {
	xmlFile, err := os.Open(fullFilePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil, err
	}
	defer xmlFile.Close()

	bytes, _ := ioutil.ReadAll(xmlFile)

	var suite TestSuite
	err = xml.Unmarshal(bytes, &suite)
	if err != nil {
		return nil, err
	}

	return &suite, nil
}
