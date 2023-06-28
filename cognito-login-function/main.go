package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	util "github.com/mildnl/congregation-noticeboard-backend/util"
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

type AuthResult struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	IdToken          string `json:"id_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type LoginResponse struct {
	Message    string     `json:"message"`
	AuthResult *cognito.AuthenticationResultType `json:"auth_result,omitempty"`
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
	var loginReq LoginRequest
	err := json.Unmarshal([]byte(request.Body), &loginReq)
	if err != nil {
		log.Println("Failed to unmarshal request body:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request payload",
		}, nil
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
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to create AWS session",
		}, err
	}

	client := cognito.New(sess)

	authTry := &cognito.InitiateAuthInput{
		AuthFlow:       flow,
		AuthParameters: params,
		ClientId:       aws.String(os.Getenv("AWS_APP_CLIENT_ID")),
	}

	res, err := client.InitiateAuth(authTry)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cognito.ErrCodeNotAuthorizedException:
				// Check if the error message indicates an expired password
				if strings.Contains(aerr.Message(), "expired and must be reset") {
					// Handle the case where the password has expired
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusBadRequest,
						Body:       "Password expired. Please reset your password.",
					}, nil
				}
			case cognito.ErrCodeUserNotFoundException:
				// Handle the case where the user does not exist
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       "User not found",
				}, nil
			case cognito.ErrCodeInvalidParameterException:
				// Handle the case where the password is invalid
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       "Invalid password",
				}, nil
			case cognito.ErrCodeUserLambdaValidationException:
				log.Printf("Lambda validation failed: %s", aerr.Message())
				// Handle the case where the password is invalid
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       fmt.Sprintf("Lambda validation failed: %s", aerr.Message()),
				}, nil
			}
			log.Println("Authentication failed:", aerr)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf("Authentication failed: %s", aerr.Message()),
			}, nil
		}
		log.Println("Failed to initiate auth:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Failed to initiate auth",
		}, nil
	}

	response := LoginResponse{
		Message:    "Authentication successful",
		AuthResult: res.AuthenticationResult,
	}
	fmt.Println(response)

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Println("Failed to marshal response:", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("failed to marshal response: %v", err)
	}

	token, err := util.GenerateAccessToken()
	if err != nil {
		log.Println("Failed to generate access token:", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("failed to generate access token: %v", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"auth": token,
			}, 
		Body: string(responseJSON),
		}, nil
}

func main() {
	lambda.Start(Handler)
}
