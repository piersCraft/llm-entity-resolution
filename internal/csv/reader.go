package csv

import (
	"encoding/csv"
	"os"

	"github.com/gocarina/gocsv"
)

func ReadInputRecords(filePath string) ([]*domain.InputRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []*domain.InputRecord
	if err := gocsv.UnmarshalFile(file, &records); err != nil {
		return nil, err
	}

	return records, nil
}
