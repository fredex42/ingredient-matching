package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

func makeBasePrompt(references []*DensityReferenceRecord) string {
	prompt := `Your job is to cross-reference ingredients between two lists.  You will be presented with an unknown ingredient and
	you must choose the best matching ingredient from the reference list. If there is no good match, you should respond with "NO MATCH".

	When choosing a match, consider that ingredients may be described in different ways. For example, "chopped tomatoes" and
	"tomato, chopped" should be considered a match. However, "tomato sauce" and "tomato paste" are different ingredients and
	should not be considered a match for "tomatoes".

	You must ONLY use the ingredients in the reference list to make your match.  Do NOT attempt to use any external knowledge.
	
	Your response should be in the form "Confidence: <confidence>, Match <ingredient>" where <confidence> is one of "HIGH", "MEDIUM", or "LOW". If there is no good match, respond with "NO MATCH".`

	prompt += "\n\nHere is the reference list:\n"
	for _, ref := range references {
		prompt += "- " + ref.Normalised + "\n"
	}

	return prompt
}

type StructuredResponse struct {
	Confidence string  `json:"confidence"`
	MatchTo    *string `json:"match_to"`
}

func ParseResponse(response *ClaudeMessageResponse) (*StructuredResponse, error) {
	parser := regexp.MustCompile(`\s*(HIGH|MEDIUM|LOW)\s*,\s*Match\s*(.+)\s*`)

	var structuredResponse StructuredResponse
	if response == nil || len(response.Content) == 0 || response.Content[0].Text == nil {
		return nil, fmt.Errorf("empty response content")
	}
	content := response.Content[0].Text

	matches := parser.FindStringSubmatch(*content)
	if len(matches) == 0 {
		if *content == "NO MATCH" || strings.HasSuffix(*content, "NO MATCH") {
			structuredResponse.Confidence = "NO MATCH"
			structuredResponse.MatchTo = nil
			return &structuredResponse, nil
		}
		return nil, fmt.Errorf("could not parse response: %s", *content)
	}

	structuredResponse.Confidence = matches[1]
	matchTo := matches[2]
	structuredResponse.MatchTo = &matchTo

	return &structuredResponse, nil
}

func FindReferenceIngredient(references []*DensityReferenceRecord, name string) *DensityReferenceRecord {
	for _, ref := range references {
		if ref.Normalised == name {
			return ref
		}
	}
	return nil
}

func ProcessRecords(brClient *bedrockruntime.Client, modelId *string, references []*DensityReferenceRecord, missing []*MissingDensitiesRecord, limit int) {
	basePrompt := makeBasePrompt(references)

	for i, record := range missing {
		if limit > 0 && i >= limit {
			break
		}

		prompt := basePrompt + "\n\nWhat is the best match for the ingredient: " + record.Ingredient + "?"

		messageStream := []*ClaudeMessage{
			NewClaudeUserMessage(prompt),
			NewClaudeAssistantMessage("Confidence:"),
		}

		for attempts := 0; attempts < 3; attempts++ {
			request := &ClaudeRequest{
				AnthropicVersion:  "bedrock-2023-05-31",
				Messages:          messageStream,
				MaxTokensToSample: aws.Int(10000),
				Temperature:       aws.Float32(0),
			}

			response, err := SendClaudeRequest(context.Background(), brClient, modelId, request)
			if err != nil {
				fmt.Printf("Error processing record %d (%s): %v\n", i+1, record.Ingredient, err)
				break
			}

			structuredResult, err := ParseResponse(response)
			if err != nil {
				fmt.Printf("Error parsing response for record %d (%s): %v\n", i+1, record.Ingredient, err)
				messageStream = append(messageStream, &ClaudeMessage{
					Role:    ROLE_ASSISTANT,
					Content: response.Content,
				})
				messageStream = append(messageStream, NewClaudeUserMessage("Your response could not be parsed. Please respond in the format 'Confidence: <confidence>, Match <ingredient>' or 'NO MATCH'."))
				continue
			}

			if structuredResult.MatchTo != nil {
				maybeReference := FindReferenceIngredient(references, *structuredResult.MatchTo)
				if maybeReference != nil {
					record.MatchTo = structuredResult.MatchTo
					record.Density = aws.Float32(maybeReference.Density)
					fmt.Printf("Record %d (%s): Matched to %s with density %f (confidence: %s)\n", i+1, record.Ingredient, *record.MatchTo, *record.Density, structuredResult.Confidence)
					break
				} else {
					fmt.Printf("Record %d (%s): Claimed match to %s not found in references\n", i+1, record.Ingredient, *structuredResult.MatchTo)
					messageStream = append(messageStream, &ClaudeMessage{
						Role:    ROLE_ASSISTANT,
						Content: response.Content,
					})
					messageStream = append(messageStream, NewClaudeUserMessage(fmt.Sprintf("The ingredient '%s' does not appear in the reference list.  Please try again.", *structuredResult.MatchTo)))
				}
			} else {
				fmt.Printf("Record %d (%s): No match found\n", i+1, record.Ingredient)
				break
			}
		}
	}
}
