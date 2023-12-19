package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkoukk/tiktoken-go"
)

type SQSMessageBody struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Model      string `json:"model"`
	User       string `json:"user"`
	Created    string `json:"created"`
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (events.APIGatewayProxyResponse, error) {
	for _, message := range sqsEvent.Records {
		var msgBody SQSMessageBody
		err := json.Unmarshal([]byte(message.Body), &msgBody)
		if err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		tke, err := tiktoken.EncodingForModel(msgBody.Model)
		if err != nil {
			fmt.Printf("getEncoding: %v\n", err)
			continue
		}

		promptTokens := len(tke.Encode(msgBody.Prompt, nil, nil))
		completionTokens := len(tke.Encode(msgBody.Completion, nil, nil))

		promptCost := 0.0
		completionCost := 0.0
		totalCost := 0.0

		baseRatio := 2.0

		switch msgBody.Model {
		case "gpt-4-vision-preview":
			promptCost = float64(promptTokens) * 0.01 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.03 * baseRatio / 1000
		case "gpt-4-1106-preview":
			promptCost = float64(promptTokens) * 0.01 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.03 * baseRatio / 1000
		case "gpt-4-0314":
			promptCost = float64(promptTokens) * 0.03 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.06 * baseRatio / 1000
		case "gpt-4":
			promptCost = float64(promptTokens) * 0.03 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.06 * baseRatio / 1000
		case "gpt-3.5-turbo-0301":
			promptCost = float64(promptTokens) * 0.0016 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
		case "gpt-3.5-turbo":
			promptCost = float64(promptTokens) * 0.0016 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
		case "gpt-3.5-turbo-16k":
			promptCost = float64(promptTokens) * 0.003 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.004 * baseRatio / 1000
		case "gpt-3.5-turbo-1106":
			promptCost = float64(promptTokens) * 0.001 * baseRatio / 1000
			completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
		default:
			fmt.Println("Model not supported")
		}

		totalCost = promptCost + completionCost

		fmt.Println("user: " + msgBody.User)
		fmt.Printf("totalTokens: %d\n", promptTokens+completionTokens)
		fmt.Printf("totalCost: %f\n", totalCost)
	}

	// Return a successful response
	return events.APIGatewayProxyResponse{Body: "OK", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
