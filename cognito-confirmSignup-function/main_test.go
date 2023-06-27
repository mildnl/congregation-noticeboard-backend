package main

import (
	"context"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

type mockCognitoClient struct{}

func (m *mockCognitoClient) ConfirmSignUp(input *cognito.ConfirmSignUpInput) (*cognito.ConfirmSignUpOutput, error) {
	// Simulate an expired validation code error
	if *input.ConfirmationCode == "expired" {
		return nil, awserr.New(cognito.ErrCodeExpiredCodeException, "Validation code expired", errors.New("expired validation code"))
	}

	// Simulate other errors
	return nil, errors.New("An unknown error occurred")
}

func TestHandler_ValidConfirmationCode(t *testing.T) {
	// Prepare a valid confirmation code request
	requestBody := `{"username": "testuser", "confirmation_code": "123456"}`
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := Handler(context.Background(), request)

	// Check the response
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	// Check the response body
	expectedBody := `{"message":"User signup confirmed"}`
	if response.Body != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, response.Body)
	}
}

func TestHandler_ExpiredConfirmationCode(t *testing.T) {
	// Prepare an expired confirmation code request
	requestBody := `{"username": "testuser", "confirmation_code": "expired"}`
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := Handler(context.Background(), request)

	// Check the response
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", response.StatusCode)
	}

	// Check the response body
	expectedBody := "Validation code expired"
	if response.Body != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, response.Body)
	}
}

func TestHandler_InvalidRequestBody(t *testing.T) {
	// Prepare an invalid request body
	requestBody := `{"invalid": "data"}`
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := Handler(context.Background(), request)

	// Check the response
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", response.StatusCode)
	}

	// Check the error message in the response body
	expectedBody := "json: unknown field \"invalid\""
	if response.Body != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, response.Body)
	}
}

func TestHandler_OtherErrors(t *testing.T) {
	// Prepare a request with an unknown username
	requestBody := `{"username": "unknownuser", "confirmation_code": "123456"}`
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := Handler(context.Background(), request)

	// Check the response
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if response.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", response.StatusCode)
	}

	// Check the response body
	expectedBody := "An unknown error occurred"
	if response.Body != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, response.Body)
	}
}
