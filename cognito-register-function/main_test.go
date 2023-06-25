package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	util "github.com/mildnl/congregation-noticeboard-backend/util"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	// Load environment variables from .env file
	err := godotenv.Load()
	assert.NoError(t, err, "Error loading .env file")

	// Prepare a sample request body
	requestBody := `{
		"family_name": "Test",
		"given_name": "User",
		"phone_number": "1234567890",
		"username": "testuser",
		"password": "",
		"email": "test@example.com"
	}`

	// Generate a random password
	password := util.GeneratePassword()

	// Update the request body with the generated password
	requestBody = strings.Replace(requestBody, `""`, `"`+password+`"`, 1)

	// Create a session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	assert.NoError(t, err, "Error creating AWS session")

	// Create a CognitoIdentityProvider client
	client := cognito.New(sess)

	// Create a sample request event
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the Lambda handler
	response, err := Handler(context.Background(), request)

	// Check for errors
	assert.NoError(t, err, "Handler returned an error")

	// Check the response status code
	assert.Equal(t, 200, response.StatusCode, "Unexpected status code")

	// Parse the response body
	var responseBody map[string]string
	err = json.Unmarshal([]byte(response.Body), &responseBody)

	// Check for JSON parsing errors
	assert.NoError(t, err, "Error parsing response body")

	// Check the response message
	expectedMessage := "User registration successful"
	assert.Equal(t, expectedMessage, responseBody["message"], "Unexpected response message")

	// Delete the user
	deleteUserInput := &cognito.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("AWS_USER_POOL_ID")),
		Username:   aws.String("testuser"),
	}
	_, err = client.AdminDeleteUser(deleteUserInput)
	assert.NoError(t, err, "Error deleting the user")
}

func TestMain(m *testing.M) {
	// Run the test case
	code := m.Run()

	// Perform teardown actions here (if any)
	os.Exit(code)
}
