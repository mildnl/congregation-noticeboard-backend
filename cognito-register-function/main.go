package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/joho/godotenv"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	family_name string `json:"family_name"`
	given_name string `json:"given_name"`
	phone_number string `json:"phone_number"`
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
	var user User
	err := json.Unmarshal([]byte(request.Body), &user)
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

	// Register the user
	input := &cognito.SignUpInput{
		ClientId: aws.String(os.Getenv("AWS_APP_CLIENT_ID")),
		Username: aws.String(user.Username),
		Password: aws.String(user.Password),
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
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Return a successful response
	response := map[string]string{
		"message": "User registration successful",
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
