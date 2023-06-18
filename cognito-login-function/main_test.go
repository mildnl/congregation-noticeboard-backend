package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"github.com/joho/godotenv"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/stretchr/testify/assert"
)

type mockCognitoClient struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.InitiateAuthOutput
	Err      error
}

func (m *mockCognitoClient) InitiateAuthWithContext(ctx aws.Context, input *cognitoidentityprovider.InitiateAuthInput, opts ...request.Option) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return m.Response, m.Err
}

func TestMain(m *testing.M) {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	// Run the tests
	exitCode := m.Run()

	// Perform any necessary test cleanup or teardown here

	// Exit with the appropriate exit code
	os.Exit(exitCode)
}

func TestLogin(t *testing.T) {
	// Create a sample login request
	loginReq := LoginRequest{
		Username:         os.Getenv("USERNAME"),
		Password:         os.Getenv("PASSWORD"),
	}

	// Marshal the login request to JSON
	reqJSON, _ := json.Marshal(loginReq)

	// Create a sample API Gateway Proxy request
	apiRequest := events.APIGatewayProxyRequest{
		Body: string(reqJSON),
	}

	// Create a context
	ctx := context.Background()

	// Create a mock response from InitiateAuth
	mockResponse := &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken:  aws.String("mockAccessToken"),
			ExpiresIn:    aws.Int64(3600),
			RefreshToken: aws.String("mockRefreshToken"),
			TokenType:    aws.String("Bearer"),
		},
	}

	// Create a mock Cognito client
	mockClient := &mockCognitoClient{
		Response: mockResponse,
	}

	// Invoke the Login function
	response, err := Login(ctx, apiRequest, mockClient)

	// Assertions
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.StatusCode, "Expected status code 200")
	assert.Equal(t, "Authentication successful", response.Body, "Expected success message")
}
