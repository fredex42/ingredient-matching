package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

func main() {
	region := flag.String("region", "eu-west-1", "The AWS region")
	model := flag.String("model", "", "The Bedrock model to use")
	reference := flag.String("reference", "density_reference.csv", "Path to density reference CSV file")
	missing := flag.String("missing", "missing_ingredients.csv", "Path to missing ingredients CSV file")
	limit := flag.Int("limit", 1, "Limit the number of records to process (0 for no limit)")
	out := flag.String("out", "", "Path to output CSV file for filled missing densities")
	flag.Parse()

	fmt.Printf("Using %s in %s\n", *model, *region)

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(*region))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}

	client := bedrockruntime.NewFromConfig(sdkConfig)

	referenceRecords, err := LoadDensityReferenceCSV(*reference)
	if err != nil {
		fmt.Printf("Error loading density reference CSV: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d density reference records\n", len(referenceRecords))

	missingRecords, err := LoadMissingDensitiesCSV(*missing)
	if err != nil {
		fmt.Printf("Error loading missing densities CSV: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d missing densities records\n", len(missingRecords))

	var csvWriter *csv.Writer = nil

	if out != nil && *out != "" {
		fmt.Printf("Writing output to %s\n", *out)
		writer, err := os.Create(*out)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer writer.Close()

		csvWriter = csv.NewWriter(writer)
		defer csvWriter.Flush()
	}
	ProcessRecords(client, model, referenceRecords, missingRecords, *limit, csvWriter)
}
