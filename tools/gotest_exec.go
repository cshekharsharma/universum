package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// TestResult represents the JSON output from `go test -v -json`
type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"` // Field to capture elapsed time
}

func main() {
	startTime := time.Now().UnixMilli()
	cmd := exec.Command("go", "test", "./...", "-v", "-json", "-coverprofile=./coverage.txt")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	// Parse each line of the output
	dec := json.NewDecoder(&out)

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	for dec.More() {
		var result TestResult
		if err := dec.Decode(&result); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		status := ""

		// Process and print results with elapsed time
		if result.Action == "pass" && result.Test != "" {
			passedTests++
			status = "\033[1;32mPASS\033[0m"
		} else if result.Action == "fail" && result.Test != "" {
			failedTests++
			status = "\033[1;31mFAIL\033[0m"
		} else if result.Action == "skip" {
			skippedTests++
			status = "\033[1;33mSKIP\033[0m"
		} else {
			continue
		}

		fmt.Printf(">> %s: \033[36m[%.2fs]\033[0m %s/%s\n", status, result.Elapsed, result.Package, result.Test)

		if failedTests > 0 {
			os.Exit(1)
		}
	}

	totalTests = passedTests + failedTests + skippedTests

	passedPercent := fmt.Sprintf("%.2f", float64(passedTests)/float64(totalTests)*100)
	failedPercent := fmt.Sprintf("%.2f", float64(failedTests)/float64(totalTests)*100)
	skippedPercent := fmt.Sprintf("%.2f", float64(skippedTests)/float64(totalTests)*100)

	fmt.Printf("=============================================================\n\n")
	fmt.Printf("\033[1;32mPASSED:  \033[0m %d/%d \t[ %v%% ]\n", passedTests, totalTests, passedPercent)
	fmt.Printf("\033[1;31mFAILED:  \033[0m %d/%d \t[ %v%% ]\n", failedTests, totalTests, failedPercent)
	fmt.Printf("\033[1;33mSKIPPED: \033[0m %d/%d \t[ %v%% ]\n\n", skippedTests, totalTests, skippedPercent)
	fmt.Printf("\033[1;36mDURATION: \033[0m \033[1;32m\u2605\u2605\u2605\033[0m %.3f seconds\n", float64(time.Now().UnixMilli()-startTime)/1000)
	fmt.Printf("=============================================================\n\n")
}
