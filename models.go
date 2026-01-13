package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type DensityReferenceRecord struct {
	Id         int     `json:"id"`
	Ingredient string  `json:"ingredient"`
	Normalised string  `json:"normalised_form"`
	Density    float32 `json:"density"`
	Source     string  `json:"source"`
}

func (r *DensityReferenceRecord) to_csv() []string {
	return []string{
		strconv.Itoa(r.Id),
		r.Ingredient,
		r.Normalised,
		fmt.Sprintf("%f", r.Density),
		r.Source,
	}
}

func ParseDensityValue(value string) (float32, error) {
	isRange := regexp.MustCompile(`^\s*([\d.]+)\s*-\s*([\d.]+)\s*$`)
	singleValue := regexp.MustCompile(`^\s*([\d.]+)\s*$`)

	if isRange.MatchString(value) {
		matches := isRange.FindStringSubmatch(value)
		if len(matches) != 3 {
			return 0, fmt.Errorf("could not parse range: %s", value)
		}
		parsedValue, err := strconv.ParseFloat(matches[2], 32)
		return float32(parsedValue), err
	} else if singleValue.MatchString(value) {
		matches := singleValue.FindStringSubmatch(value)
		if len(matches) != 2 {
			return 0, fmt.Errorf("could not parse value: %s", value)
		}
		parsedValue, err := strconv.ParseFloat(matches[1], 32)
		if err != nil {
			return 0, err
		}
		return float32(parsedValue), nil
	} else {
		return 0, fmt.Errorf("could not parse value: %s", value)
	}
}

func ParseDensityReferenceRecord(row []string) (*DensityReferenceRecord, error) {
	if len(row) < 5 {
		return nil, fmt.Errorf("unexpected number of fields: %d", len(row))
	}

	var id int
	var err error
	if row[0] == "" {
		id = -1
	} else {
		id, err = strconv.Atoi(row[0])
		if err != nil {
			return nil, fmt.Errorf("invalid id: %v", err)
		}
	}
	density, err := ParseDensityValue(row[3])
	if err != nil {
		return nil, fmt.Errorf("invalid density: %v", err)
	}

	return &DensityReferenceRecord{
		Id:         id,
		Ingredient: row[1],
		Normalised: row[2],
		Density:    float32(density),
		Source:     row[4],
	}, nil
}

type MissingDensitiesRecord struct {
	Popularity int      `json:"popularity"`
	Ingredient string   `json:"density_ingredient"`
	Action     *string  `json:"action"`
	MatchTo    *string  `json:"match_to"`
	Density    *float32 `json:"density"`
	Example    string   `json:"example"`
	Confidence string   `json:"confidence"`
}

// Convert a single record to CSV row
func (r *MissingDensitiesRecord) to_csv() []string {
	action := ""
	if r.Action != nil {
		action = *r.Action
	}
	matchTo := ""
	if r.MatchTo != nil {
		matchTo = *r.MatchTo
	}
	density := ""
	if r.Density != nil {
		density = fmt.Sprintf("%f", *r.Density)
	}
	return []string{
		strconv.Itoa(r.Popularity),
		r.Ingredient,
		action,
		matchTo,
		density,
		r.Example,
		r.Confidence,
	}
}

// Parse a CSV row into a record
func ParseMissingDensitiesRecord(row []string) (*MissingDensitiesRecord, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("unexpected number of fields: %d", len(row))
	}

	popularity, err := strconv.Atoi(row[0])
	if err != nil {
		return nil, fmt.Errorf("invalid popularity: %v", err)
	}

	var density *float32 = nil

	if row[4] != "" {
		densityValue, err := strconv.ParseFloat(row[4], 32)
		if err != nil {
			return nil, fmt.Errorf("invalid density: %v", err)
		}
		density = new(float32)
		*density = float32(densityValue)
	}

	action := &row[2]
	if row[2] == "" {
		action = nil
	}

	matchTo := &row[3]
	if row[3] == "" {
		matchTo = nil
	}

	confidence := ""
	if len(row) > 6 {
		confidence = row[6]
	}

	return &MissingDensitiesRecord{
		Popularity: popularity,
		Ingredient: row[1],
		Action:     action,
		MatchTo:    matchTo,
		Density:    density,
		Example:    row[5],
		Confidence: confidence,
	}, nil
}
