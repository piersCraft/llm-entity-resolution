package config

type Config struct {
	InputFile   string
	OutputFile  string
	APIEndpoint string
	APIToken    string
	Workers     int // Number of concurrent workers
}

func LoadConfig() *Config {
	return &Config{
		InputFile:   "input.csv",
		OutputFile:  "output.csv",
		APIEndpoint: "https://api.example.com/process",
		APIToken:    "your-api-token",
		Workers:     5,
	}
}
