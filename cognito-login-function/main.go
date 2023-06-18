package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/joho/godotenv"
)

const flowUsernamePassword = "USER_PASSWORD_AUTH"
const flowRefreshToken = "REFRESH_TOKEN_AUTH"

type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Refresh      string `json:"refresh"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponse struct {
	Message    string                            `json:"message"`
	AuthResult *cognito.AuthenticationResultType `json:"auth_result"`
}

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
}

func Login(ctx context.Context, request events.APIGatewayProxyRequest, cognitoClient cognitoidentityprovideriface.CognitoIdentityProviderAPI) (events.APIGatewayProxyResponse, error) {
	// Parse the request body
	var loginReq LoginRequest
	err := json.Unmarshal([]byte(request.Body), &loginReq)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
	}

	flow := aws.String(flowUsernamePassword)
	params := map[string]*string{
		"USERNAME": aws.String(loginReq.Username),
		"PASSWORD": aws.String(loginReq.Password),
	}

	if loginReq.Refresh != "" {
		flow = aws.String(flowRefreshToken)
		params = map[string]*string{
			"REFRESH_TOKEN": aws.String(loginReq.RefreshToken),
		}
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to create AWS session: %v", err)
	}

	client := cognito.New(sess)

	authTry := &cognito.InitiateAuthInput{
		AuthFlow:       flow,
		AuthParameters: params,
		ClientId:       aws.String(os.Getenv("AWS_APP_CLIENT_ID")),
	}

	res, err := client.InitiateAuth(authTry)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, fmt.Errorf("failed to initiate auth: %v", err)
	}

	response := LoginResponse{
		Message:    "Authentication successful",
		AuthResult: res.AuthenticationResult,
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: response.Message}, nil
}

func main() {
	lambda.Start(Login)
}
