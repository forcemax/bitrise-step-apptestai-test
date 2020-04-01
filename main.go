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
	"strconv"
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

type structDataOfExecuteTest struct {
	TestCount int `json:"test_count"`
	TestsetID int `json:"testset_id"`
}
type structExecuteTest struct {
	Data       structDataOfExecuteTest `json:"data"`
	Reason     string                  `json:"reason"`
	ResultCode int                     `json:"result_code"`
	ResultMsg  string                  `json:"result_msg"`
}

type structTestsetStatusDetail struct {
	ErrorCnt        int `json:"error_cnt"`
	FailCnt         int `json:"fail_cnt"`
	InitializingCnt int `json:"initializing_cnt"`
	PassCnt         int `json:"pass_cnt"`
	RunningCnt      int `json:"running_cnt"`
	StopCnt         int `json:"stop_cnt"`
	TotalTestCnt    int `json:"total_test_cnt"`
}
type structDataOfCheckComplete struct {
	ResponseTime        string                    `json:"response_time"`
	TestsetStatusDetail structTestsetStatusDetail `json:"testset_status_detail"`
	TestsetStatus       string                    `json:"testset_status"`
}
type structCheckComplete struct {
	Data       structDataOfCheckComplete `json:"data"`
	Reason     string                    `json:"reason"`
	ResultCode int                       `json:"result_code"`
	ResultMsg  string                    `json:"result_msg"`
}

type structDataOfGetTestResult struct {
	Complete   bool   `json:"complete"`
	ResultHTML string `json:"result_html"`
	ResultJSON string `json:"result_json"`
	ResultXML  string `json:"result_xml"`
}
type structGetTestResult struct {
	Data       structDataOfGetTestResult `json:"data"`
	Reason     string                    `json:"reason"`
	ResultCode int                       `json:"result_code"`
	ResultMsg  string                    `json:"result_msg"`
}

type structCredentials struct {
	LoginID string `json:"login_id"`
	LoginPW string `json:"login_pw"`
}
type structParams struct {
	TestsetName string            `json:"testset_name"`
	TimeLimit   int               `json:"time_limit"`
	UseVO       bool              `json:"use_vo"`
	Callback    string            `json:"callback"`
	Credentials structCredentials `json:"credentials"`
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

func multipartRequest(uri string, data string, paramName, path string) (*http.Request, error) {
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

	_ = writer.WriteField("data", data)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, err
}

func executeTest(accesskey string, projectid string, packagefile string, params structParams) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	var data = ""
	data += "{\"pid\": " + projectid
	data += ", \"testset_name\": \"" + params.TestsetName + "\""
	if params.TimeLimit >= 5 && params.TimeLimit <= 30 {
		data += ", \"time_limit\": " + strconv.Itoa(params.TimeLimit)
	}
	data += ", \"use_vo\": " + strconv.FormatBool(params.UseVO)
	if len(params.Callback) > 0 {
		data += ", \"callback\": \"" + params.Callback + "\""
	}
	if len(params.Credentials.LoginID) > 0 && len(params.Credentials.LoginPW) > 0 {
		data += ", \"credentials\": { \"login_id\": \"" + params.Credentials.LoginID + "\", \"login_pw\": \"" + params.Credentials.LoginPW + "\"}"
	}
	data += "}"

	// data := "{\"pid\":" + projectid + ", \"test_set_name\":\"" + testsetname + "\"}"
	request, err := multipartRequest("https://api.apptest.ai/openapi/v2/testset", data, "app_file", packagefile)
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

func checkComplete(accesskey string, tsID string) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	url := "https://api.apptest.ai/openapi/v2/testset/" + tsID

	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(authToken[0], authToken[1])

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
		panic("Check complete failed.")
	} else {
		if resp.StatusCode != 200 {
			panic("Check complete failed : HTTP status code " + fmt.Sprintf("%d", resp.StatusCode))
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return string(body)
		}
	}
}

func getTestResult(accesskey string, tsID string) (bodyContent string) {
	authToken := strings.Split(accesskey, ":")
	url := "https://api.apptest.ai/openapi/v2/testset/" + tsID + "/result"

	request, _ := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(authToken[0], authToken[1])

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
		panic("Get result failed.")
	} else {
		if resp.StatusCode != 200 {
			panic("Get result failed : HTTP status code " + fmt.Sprintf("%d", resp.StatusCode))
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

	testsetname := os.Getenv("testset_name")
	timelimit, err := strconv.Atoi(os.Getenv("time_limit"))
	if err != nil {
		fmt.Println("time_limit value is invalid. follows the time-limit saved in the project.")
		timelimit = 0
	}
	usevo, err := strconv.ParseBool(os.Getenv("use_vo"))
	if err != nil {
		fmt.Println("use_vo value is invalid. use_vo set default value(false).")
		usevo = false
	}
	callback := os.Getenv("callback")
	loginid := os.Getenv("login_id")
	loginpw := os.Getenv("login_pw")

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

	if len(testsetname) == 0 {
		testsetname = os.Getenv("BITRISE_GIT_MESSAGE")
		if len(testsetname) != 0 {
			testsetname = clearCommitMessage(testsetname)
		} else {
			testsetname = os.Getenv("BITRISE_GIT_COMMIT")
		}
		if len(testsetname) != 0 {
			testsetname = "no commit message"
		}
	}

	var params structParams
	params.TestsetName = testsetname
	params.TimeLimit = timelimit
	params.UseVO = usevo
	params.Callback = callback
	params.Credentials.LoginID = loginid
	params.Credentials.LoginPW = loginpw

	responseBody := executeTest(accesskey, projectid, binarypath, params)
	if len(responseBody) == 0 {
		panic("Test initiation failed: no response.")
	}
	var ret structExecuteTest
	if err := json.Unmarshal([]byte(responseBody), &ret); err != nil {
		panic("Test initiation failed: " + string(err.Error()))
	}

	if ret.Data.TestsetID == 0 {
		panic("Test initialize failed: invalid response.")
	}
	tsID := fmt.Sprintf("%d", ret.Data.TestsetID)
	fmt.Println("Test initiated.")

	stepCount := 0
	for running == true {
		time.Sleep(15 * time.Second)
		stepCount++
		fmt.Println("Test is progressing... " + fmt.Sprintf("%d", stepCount*15) + "sec.")

		responseBody := checkComplete(accesskey, tsID)
		if len(responseBody) == 0 {
			panic("Test progress check failed : no response")
		}
		var ret structCheckComplete
		if err := json.Unmarshal([]byte(responseBody), &ret); err != nil {
			panic("Test progress check failed: " + string(err.Error()))
		}
		if ret.Data.TestsetStatus == "Complete" {
			responseBody := getTestResult(accesskey, tsID)
			if len(responseBody) == 0 {
				panic("Test result get failed : no response")
			}
			var ret structGetTestResult
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
