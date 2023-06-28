package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/joho/godotenv"
)

type UserInformation struct {
	Username 		string `json:"username"`
	Email    		string `json:"email"`
	AccessToken    	string `json:"access_token"`
	// Add other user attributes here as needed
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
	var userInformation UserInformation
	err := json.Unmarshal([]byte(request.Body), &userInformation)
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

	// Get user information
	input := &cognito.GetUserInput{
		AccessToken: aws.String(userInformation.AccessToken),
	}

	result, err := client.GetUser(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Extract the required user attributes
	username := aws.StringValue(result.Username)
	email := aws.StringValue(result.UserAttributes[0].Value)
	// Add other user attribute extractions here as needed

	// Prepare the response
	response := UserInformation{
		Username: username,
		Email:    email,
		// Assign other extracted user attributes here as needed
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
