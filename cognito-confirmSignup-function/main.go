package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/joho/godotenv"
)

type ConfirmationRequest struct {
	Username     string `json:"username"`
	ConfirmationCode string `json:"confirmation_code"`
}

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse the request body
	var confirmationRequest ConfirmationRequest
	err := json.Unmarshal([]byte(request.Body), &confirmationRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	// Create a session
	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Create a CognitoIdentityProvider client
	client := cognito.New(sess)

	// Confirm the user's signup
	input := &cognito.ConfirmSignUpInput{
		ClientId:         aws.String(os.Getenv("AWS_APP_CLIENT_ID")),
		Username:         aws.String(confirmationRequest.Username),
		ConfirmationCode: aws.String(confirmationRequest.ConfirmationCode),
	}
	
	_, err = client.ConfirmSignUp(input)
	if err != nil {
		// Check if the error is due to an expired validation code
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == cognito.ErrCodeExpiredCodeException {
				// Handle the expired code error
				return events.APIGatewayProxyResponse{
					StatusCode: 400,
					Body:       "Validation code expired",
				}, nil
			}
		}
	
		// Handle other errors
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}


	// Return a successful response
	response := map[string]string{
		"message": "User signup confirmed",
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
