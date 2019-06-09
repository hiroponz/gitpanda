package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if isPrintVersion {
		printVersion()
		return
	}

	checkEnv("GITLAB_API_ENDPOINT")
	checkEnv("GITLAB_BASE_URL")
	checkEnv("GITLAB_PRIVATE_TOKEN")
	checkEnv("SLACK_OAUTH_ACCESS_TOKEN")

	if os.Getenv("LAMBDA_TASK_ROOT") != "" && os.Getenv("LAMBDA_RUNTIME_DIR") != "" {
		// for AWS Lambda
		lambda.Start(lambdaHandler)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("gitpanda started: port=%s\n", port)
	http.HandleFunc("/", normalHandler)
	http.ListenAndServe(":"+port, nil)
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := strings.TrimSpace(request.Body)

	s := NewSlackWebhook(
		os.Getenv("SLACK_OAUTH_ACCESS_TOKEN"),
		&GitLabURLParserParams{
			APIEndpoint:  os.Getenv("GITLAB_API_ENDPOINT"),
			BaseURL:      os.Getenv("GITLAB_BASE_URL"),
			PrivateToken: os.Getenv("GITLAB_PRIVATE_TOKEN"),
		},
	)
	response, err := s.Request(
		body,
		false,
	)

	if err != nil {
		log.Printf("[ERROR] body=%s, response=%s, error=%v", body, response, err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       response,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       response,
	}, nil
}
