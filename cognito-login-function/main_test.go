package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/mildnl/congregation-noticeboard-backend/util"
	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	family_name string `json:"family_name"`
	given_name string `json:"given_name"`
	phone_number string `json:"phone_number"`
}

var testUserPassword string

type mockCognitoClient struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.InitiateAuthOutput
	Err      error
}

func (m *mockCognitoClient) InitiateAuthWithContext(ctx aws.Context, input *cognitoidentityprovider.InitiateAuthInput, opts ...request.Option) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return m.Response, m.Err
}

func TestMain(m *testing.M) {
	// setup
	setup()

	// Run the tests
	exitCode := m.Run()

	// Perform any necessary test cleanup or teardown here
	teardown()

	// Exit with the appropriate exit code
	os.Exit(exitCode)
}

func setup() {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

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
	testUserPassword = util.GeneratePassword()

	// Update the request body with the generated password
	requestBody = strings.Replace(requestBody, `""`, `"`+testUserPassword+`"`, 1)

	// Parse the request body
	var user TestUser
	err = json.Unmarshal([]byte(requestBody), &user)
	if err != nil {
		log.Fatalf("Error parsing request body: %s", err.Error())
		return
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("Error creating session: %s", err.Error())
		return
	}

	// Create a CognitoIdentityProvider client
	client := cognito.New(sess)

	// Register the user
	input := &cognito.SignUpInput{
		ClientId: aws.String(os.Getenv("AWS_APP_CLIENT_ID")),
		Username: aws.String(user.Username),
		Password: aws.String(testUserPassword),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(user.Email),
			},
			{
				Name:  aws.String("family_name"),
				Value: aws.String(user.family_name),
			},
			{
				Name:  aws.String("given_name"),
				Value: aws.String(user.given_name),
			},
			{
				Name:  aws.String("phone_number"),
				Value: aws.String(user.phone_number),
			},
		},
	}

	_, err = client.SignUp(input)
	if err != nil {
		log.Fatalf("Error registering user: %s", err.Error())
		return
	}

	// Initiate user confirmation
	confirmationInput := &cognito.AdminConfirmSignUpInput{
		UserPoolId: aws.String(os.Getenv("AWS_USER_POOL_ID")),
		Username:   aws.String(user.Username),
	}

	_, err = client.AdminConfirmSignUp(confirmationInput)
	if err != nil {
		log.Fatalf("Error confirming user signup: %s", err.Error())
		return
	}
}


func teardown() {
	// Do any teardown here
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("Error creating session: %s", err.Error())
		return
	}

	// Create a CognitoIdentityProvider client
	client := cognito.New(sess)

	// Delete the user
	deleteUserInput := &cognito.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("AWS_USER_POOL_ID")),
		Username:   aws.String("testuser"),
	}
	_, err = client.AdminDeleteUser(deleteUserInput)
	if err != nil {
		log.Fatalf("Error deleting user: %s", err.Error())
		return
	}
}


func TestLogin(t *testing.T) {
	// Create a sample login request
	loginReq := LoginRequest{
		Username:         "testuser",
		Password:         testUserPassword,
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
	response, err := Handler(ctx, apiRequest, mockClient)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	}

	// Extract the body from the APIGatewayProxyResponse
	body := []byte(response.Body)

	// Create an instance of the LoginResponse struct
	var loginResponse LoginResponse

	// Unmarshal the JSON into the response object
	err = json.Unmarshal([]byte(body), &loginResponse)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	}

	// Assertions
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.StatusCode, "Expected status code 200")
	assert.Equal(t, "Authentication successful", loginResponse.Message, "Expected success message")
}
