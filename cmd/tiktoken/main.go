package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkoukk/tiktoken-go"
	"os"
	"strconv"
	"sync"
)

const (
	defaultBaseRatio = 1.0
)

var (
	baseRatio float64
)

func init() {
	// Load the base ratio once during the initialization of the Lambda.
	envBaseRatio := os.Getenv("BASE_RATIO")
	if envBaseRatio == "" {
		baseRatio = defaultBaseRatio
	} else {
		var err error
		baseRatio, err = strconv.ParseFloat(envBaseRatio, 64)
		if err != nil {
			fmt.Printf("Error parsing BASE_RATIO: %v, using default value.\n", err)
			baseRatio = defaultBaseRatio
		}
	}
}

type MsgBody struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Model      string `json:"model"`
}

func calculateCost(model string, promptTokens, completionTokens int) (promptCost, completionCost float64) {
	costPerThousandTokens := map[string]struct {
		in  float64
		out float64
	}{
		"gpt-4-vision-preview": {0.01, 0.03},
		"gpt-4-1106-preview":   {0.01, 0.03},
		"gpt-4-0314":           {0.03, 0.06},
		"gpt-4":                {0.03, 0.06},
		"gpt-3.5-turbo-0301":   {0.0016, 0.002},
		"gpt-3.5-turbo":        {0.0016, 0.002},
		"gpt-3.5-turbo-16k":    {0.003, 0.004},
		"gpt-3.5-turbo-1106":   {0.001, 0.002},
	}

	if _1kPrice, ok := costPerThousandTokens[model]; ok {
		promptCost = float64(promptTokens) * _1kPrice.in * baseRatio / 1000
		completionCost = float64(completionTokens) * _1kPrice.out * baseRatio / 1000
	} else {
		fmt.Println("Model not supported")
	}

	return promptCost, completionCost
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var msgBody MsgBody
	err := json.Unmarshal([]byte(event.Body), &msgBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "Error unmarshaling message: %v"}`, err),
			StatusCode: 400,
		}, nil
	}

	tke, err := tiktoken.EncodingForModel(msgBody.Model)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "getEncoding: %v"}`, err),
			StatusCode: 400,
		}, nil
	}

	// Use a WaitGroup to wait for both goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(2)

	var promptTokens int
	var completionTokens int

	// Calculate promptTokens in a separate goroutine.
	go func() {
		defer wg.Done()
		promptTokens = len(tke.Encode(msgBody.Prompt, nil, nil))
	}()

	// Calculate completionTokens in a separate goroutine.
	go func() {
		defer wg.Done()
		completionTokens = len(tke.Encode(msgBody.Completion, nil, nil))
	}()

	// Wait for both goroutines to finish.
	wg.Wait()

	promptCost, completionCost := calculateCost(msgBody.Model, promptTokens, completionTokens)
	totalCost := promptCost + completionCost

	responseBody := fmt.Sprintf(`{
		"usage": {
			"prompt_tokens": %d,
			"completion_tokens": %d,
			"total_tokens": %d
		},
		"cost": {
			"prompt_cost": %f,
			"completion_cost": %f,
			"total_cost": %f
		}
	}`, promptTokens, completionTokens, promptTokens+completionTokens, promptCost, completionCost, totalCost)

	return events.APIGatewayProxyResponse{Body: responseBody, StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
