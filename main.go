package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// for testsuite json response
type sError struct {
	Message string `json:"message"`
}
type systemOut struct {
	Contents []string `json:"contents"`
}
type testCase struct {
	Name      string    `json:"name"`
	Time      string    `json:"time"`
	SystemOut systemOut `json:"system-out"`
	Error     sError    `json:"error"`
}
type testSuite struct {
	Name      string     `json:"name"`
	TestCases []testCase `json:"testcase"`
}
type testSuites struct {
	Name      string      `json:"name"`
	TestSuite []testSuite `json:"testsuite"`
}
type testResult struct {
	TestSuites testSuites `json:"testsuites"`
}

// for json response
type sData struct {
	TestCount  int    `json:"test_count"`
	TsID       int    `json:"tsid"`
	ResultJSON string `json:"result_json"`
	ResultXML  string `json:"result_xml"`
	ResultHTML string `json:"result_html"`
}

// for test/run json response
// {
// 	data: { test_count: 3, tsid: 799632 },
// 	errorCode: 0,
// 	reason: '',
// 	result: 'ok'
// }
type executeTestResult struct {
	Data      sData  `json:"data"`
	ErrorCode int    `json:"errorCode"`
	Reason    string `json:"reason"`
	Result    string `json:"result"`
}

// for check finish json response
// { complete: false, data: {}, errorCode: 0, reason: '', result: 'ok' }
type checkCompleteResult struct {
	Complete  bool   `json:"complete"`
	Data      sData  `json:"data"`
	ErrorCode int    `json:"errorCode"`
	Reason    string `json:"reason"`
	Result    string `json:"result"`
}

// for color print
var (
	Black   = color("\033[1;30m%s\033[0m")
	Red     = color("\033[1;31m%s\033[0m")
	Green   = color("\033[1;32m%s\033[0m")
	Yellow  = color("\033[1;33m%s\033[0m")
	Purple  = color("\033[1;34m%s\033[0m")
	Magenta = color("\033[1;35m%s\033[0m")
	Teal    = color("\033[1;36m%s\033[0m")
	White   = color("\033[1;37m%s\033[0m")
)

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}
	return sprint
}

// params map[string]string,
func multipartRequest(uri string, params string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)

	_ = writer.WriteField("data", params)
	// for key, val := range params {
	// 	_ = writer.WriteField(key, val)
	// }
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, err
}

func executeTest(accesskey string, projectid string, packagefile string, testsetname string) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	params := "{\"pid\":" + projectid + ", \"test_set_name\":\"" + testsetname + "\"}"
	// extraParams := map[string]string{
	// 	"title":       "My Document",
	// 	"author":      "Matt Aimonetti",
	// 	"description": "A document with all the Go programming language secrets",
	// }
	request, err := multipartRequest("https://api.apptest.ai/openapi/v1/test/run", params, "apk_file", packagefile)
	if err != nil {
		log.Fatal(err)
	}
	request.SetBasicAuth(authToken[0], authToken[1])

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
		panic("Test initiation failed.")
	} else {
		if resp.StatusCode != 200 {
			panic("Test initiation failed : HTTP status code " + fmt.Sprintf("%d", resp.StatusCode))
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return string(body)
		}
	}
}

func checkComplete(accesskey string, projectid string, tsID string) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	url := "https://api.apptest.ai/openapi/v1/project/" + projectid + "/testset/" + tsID + "/result/all"

	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(authToken[0], authToken[1])

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
		panic("Check finish failed.")
	} else {
		if resp.StatusCode != 200 {
			panic("Check finish failed : HTTP status code " + fmt.Sprintf("%d", resp.StatusCode))
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return string(body)
		}
	}
}

func getTestResult(accesskey string, projectid string, tsID string) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	url := "https://api.apptest.ai/openapi/v1/project/" + projectid + "/testset/" + tsID + "/result/all"

	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(authToken[0], authToken[1])

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
		panic("Get result failed.")
	} else {
		if resp.StatusCode != 200 {
			panic("Get result failed : HTTP status code : " + fmt.Sprintf("%d", resp.StatusCode))
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return string(body)
		}
	}
}

func makeResult(jsonTestResult string) (outputTable string) {
	var top testResult
	if err := json.Unmarshal([]byte(jsonTestResult), &top); err != nil {
		panic(err)
	}

	outputTable = "\n"
	outputTable += "+-----------------------------------------------------------------+\n"
	outputTable += "|                        Device                        |  Result  |\n"
	outputTable += "+-----------------------------------------------------------------+\n"

	testcases := top.TestSuites.TestSuite[0].TestCases
	for _, testcase := range testcases {
		outputTable += "| " + fmt.Sprintf("%-52v", testcase.Name) + " |  "
		if len(testcase.SystemOut.Contents) > 0 {
			outputTable += Green("Passed")
		} else {
			outputTable += Red("Failed")
		}
		outputTable += "  |\n"
	}
	outputTable += "+-----------------------------------------------------------------+\n"

	return outputTable
}

func makeResultFailed(jsonTestResult string) (outputTable string) {
	var top testResult
	if err := json.Unmarshal([]byte(jsonTestResult), &top); err != nil {
		panic(err)
	}

	outputTable = "\n"
	outputTable += "+-----------------------------------------------------------------+\n"
	outputTable += "|                        Device                        |  Result  |\n"
	outputTable += "+-----------------------------------------------------------------+\n"

	testcases := top.TestSuites.TestSuite[0].TestCases
	for _, testcase := range testcases {
		if len(testcase.SystemOut.Contents) > 0 {
			continue
		}
		outputTable += "| " + fmt.Sprintf("%-52v", testcase.Name) + " |  "
		outputTable += Red("Failed")
		outputTable += "  |\n"
		outputTable += "| " + testcase.Error.Message + "\n"
	}
	outputTable += "+-----------------------------------------------------------------+\n"

	return outputTable
}

func getErrors(jsonTestResult string) (errors []testCase) {
	var top testResult
	if err := json.Unmarshal([]byte(jsonTestResult), &top); err != nil {
		panic(err)
	}

	testcases := top.TestSuites.TestSuite[0].TestCases
	for _, testcase := range testcases {
		if len(testcase.SystemOut.Contents) == 0 {
			errors = append(errors, testcase)
		}
	}

	return errors
}

func clearCommitMessage(commitMessage string) (retMessage string) {
	retMessage = commitMessage
	if len(retMessage) > 99 {
		retMessage = retMessage[0:99]
	}

	if strings.Index(retMessage, "\n") != -1 {
		retMessage = retMessage[0:strings.Index(retMessage, "\n")]
	}
	return retMessage
}

func main() {
	defer func() {
		s := recover()
		fmt.Println(s)
		os.Exit(1)
	}()

	running := true
	accesskey := os.Getenv("access_key")
	projectid := os.Getenv("project_id")
	binarypath := os.Getenv("binary_path")

	if len(accesskey) == 0 {
		panic("access_key is required parameter.")
	}

	if len(projectid) == 0 {
		panic("project_id is required parameter.")
	}

	if len(binarypath) == 0 {
		panic("binary_path is required parameter.")
	}

	if strings.Index(accesskey, ":") == -1 {
		panic("The format of access_key is incorrect.")
	}

	if _, err := os.Stat(binarypath); os.IsNotExist(err) {
		panic("binary_path file not exists.")
	}

	testsetname := os.Getenv("test_set_name")
	if len(testsetname) == 0 {
		testsetname = os.Getenv("BITRISE_GIT_MESSAGE")
		if len(testsetname) != 0 {
			testsetname = clearCommitMessage(testsetname)
		} else {
			testsetname = os.Getenv("BITRISE_GIT_COMMIT")
		}
	}

	// var tsID string
	responseBody := executeTest(accesskey, projectid, binarypath, testsetname)
	if len(responseBody) == 0 {
		panic("Test initiation failed: no response.")
	}
	var ret executeTestResult
	if err := json.Unmarshal([]byte(responseBody), &ret); err != nil {
		panic("Test initiation failed: " + string(err.Error()))
	}

	if ret.Data.TsID == 0 {
		panic("Test initialize failed: invalid response.")
	}
	tsID := fmt.Sprintf("%d", ret.Data.TsID)
	fmt.Println("Test initiated.")

	stepCount := 0
	for running == true {
		time.Sleep(15 * time.Second)
		stepCount++
		fmt.Println("Test is progressing... " + fmt.Sprintf("%d", stepCount*15) + "sec.")

		responseBody := checkComplete(accesskey, projectid, tsID)
		if len(responseBody) == 0 {
			panic("Test progress check failed : no response")
		}
		var ret checkCompleteResult
		if err := json.Unmarshal([]byte(responseBody), &ret); err != nil {
			panic("Test progress check failed: " + string(err.Error()))
		}
		if ret.Complete == true {
			responseBody := getTestResult(accesskey, projectid, tsID)
			if len(responseBody) == 0 {
				panic("Test result get failed : no response")
			}
			var ret checkCompleteResult
			if err := json.Unmarshal([]byte(responseBody), &ret); err != nil {
				panic("Test result get failed: " + string(err.Error()))
			}

			fmt.Println("Test finished.")
			outputTable := makeResult(ret.Data.ResultJSON)
			fmt.Println(outputTable)

			errors := getErrors(ret.Data.ResultJSON)
			if len(errors) > 0 {
				outputTable := makeResultFailed(ret.Data.ResultJSON)
				fmt.Println(outputTable)
				os.Exit(1)
			}
			running = false
		}
	}
	os.Exit(0)
}
