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

		promptTokens := tke.Encode(msgBody.Prompt, nil, nil)
		completionTokens := tke.Encode(msgBody.Completion, nil, nil)

		fmt.Println("prompt_tokens:", len(promptTokens))
		fmt.Println("completion_tokens:", len(completionTokens))
		fmt.Println("total_tokens:", len(promptTokens)+len(completionTokens))
	}

	// Return a successful response
	return events.APIGatewayProxyResponse{Body: "OK", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
