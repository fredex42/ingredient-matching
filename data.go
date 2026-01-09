package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func LoadDensityReferenceCSV(filePath string) ([]*DensityReferenceRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v\n", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var records []*DensityReferenceRecord

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("could not read header: %v\n", err)
	}

	ctr := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not read row %d: %v\n", ctr, err)
		}

		record, err := ParseDensityReferenceRecord(row)
		if err != nil {
			fmt.Printf("WARNING could not parse row %d: %v\n", ctr, err)
		} else {
			records = append(records, record)
		}
		ctr += 1
	}

	return records, nil
}

func SaveDensityReferenceCSV(filePath string, records []*DensityReferenceRecord) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "ingredient", "normalised_form", "density", "source"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("could not write header: %v", err)
	}

	for _, record := range records {
		if err := writer.Write(record.to_csv()); err != nil {
			return fmt.Errorf("could not write record: %v", err)
		}
	}

	return nil
}

func LoadMissingDensitiesCSV(filePath string) ([]*MissingDensitiesRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var records []*MissingDensitiesRecord

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("could not read header: %v\n", err)
	}

	ctr := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not read row %d: %v\n", ctr, err)
		}

		record, err := ParseMissingDensitiesRecord(row)
		if err != nil {
			fmt.Printf("WARNING could not parse row %d: %v\n", ctr, err)
		} else {
			records = append(records, record)
		}
		ctr += 1
	}

	return records, nil
}

func SaveMissingDensitiesCSV(filePath string, records []*MissingDensitiesRecord) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"popularity", "density_ingredient", "action", "match_to", "density", "example"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("could not write header: %v", err)
	}

	for _, record := range records {
		if err := writer.Write(record.to_csv()); err != nil {
			return fmt.Errorf("could not write record: %v", err)
		}
	}

	return nil
}
