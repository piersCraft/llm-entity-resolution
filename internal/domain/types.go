package domain

// InputRecord represents a single record from input CSV
type InputRecord struct {
	Data string `csv:"data"` // assuming single column CSV
}

// OutputRecord represents the processed result
type OutputRecord struct {
	InputData  string
	Response   string
	StatusCode int
	Success    bool
}
