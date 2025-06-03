package main

import (
	"bytes"
	// "context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Configurable constants
const (
	batchSize   = 100                         // Number of items to process concurrently
	maxRetries  = 3                           // Maximum number of retry attempts
	retryDelay  = 1 * time.Second             // Initial delay between retries
	apiEndpoint = "https://api.exa.ai/search" // Your API endpoint
	apiKey      = "5ced1c0e-f8f4-49b3-94c5-d1d091a117ef"
	timeout     = 30 * time.Second // HTTP client timeout
)

// RequestPayload represents the JSON structure sent to the API
type RequestPayload struct {
	Query      string   `json:"query"`
	Category   string   `json:"category"`
	NumResults int      `json:"numResults"`
	Contents   Contents `json:"contents"`
}

type Contents struct {
	Text bool `json:"text"`
}

// ResponsePayload represents the expected JSON response from the API
type ResponsePayload struct {
	Result  string `json:"result"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ProcessResult stores the input and corresponding API response
type ProcessResult struct {
	Input    string
	Response ResponsePayload
	Error    error
}

func main() {
	// Check command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input.csv> <output.csv>")
		os.Exit(1)
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read input CSV
	inputs, err := readCSV(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Process inputs in concurrent batches
	results := processConcurrently(inputs)

	// Write results to output CSV
	if err := writeResults(outputFile, results); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processing complete. Results written to %s\n", outputFile)
}

// Rest of the functions (readCSV, processConcurrently, writeResults) remain the same as in the original solution
// Only the callAPI function is modified as shown above

// HELPER FUNCTIONS

// Read inputs from csv
func readCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var inputs []string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV: %w", err)
		}
		// Assuming one string per line (first column)
		if len(record) > 0 {
			inputs = append(inputs, record[0])
		}
	}

	return inputs, nil
}

// API post request
func callAPI(client *http.Client, input string) (ResponsePayload, error) {
	var response ResponsePayload
	var lastErr error

	// Create payload with constant values and dynamic query
	payload := RequestPayload{
		Query:      input,
		Category:   "company",
		NumResults: 1,
		Contents: Contents{
			Text: true,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return response, fmt.Errorf("marshaling JSON: %w", err)
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * retryDelay)
		}

		req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("creating request: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("sending request: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("decoding response: %w", err)
			continue
		}
		resp.Body.Close()

		return response, nil
	}

	return response, fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
}

// Concurrent processing
func processConcurrently(inputs []string) []ProcessResult {
	var wg sync.WaitGroup
	results := make([]ProcessResult, len(inputs))
	semaphore := make(chan struct{}, batchSize) // Semaphore for limiting concurrency

	client := &http.Client{
		Timeout: timeout,
	}

	for i, input := range inputs {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(index int, input string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			result := ProcessResult{Input: input}
			result.Response, result.Error = callAPI(client, input)
			results[index] = result
		}(i, input)
	}

	wg.Wait()
	return results
}

// Write results to file
func writeResults(filename string, results []ProcessResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Input", "Result", "Success", "Error"}); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Write data rows
	for _, result := range results {
		success := "false"
		if result.Response.Success {
			success = "true"
		}

		record := []string{
			result.Input,
			result.Response.Result,
			success,
		}

		if result.Error != nil {
			record = append(record, result.Error.Error())
		} else if result.Response.Error != "" {
			record = append(record, result.Response.Error)
		} else {
			record = append(record, "")
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("writing record: %w", err)
		}
	}

	return nil
}
