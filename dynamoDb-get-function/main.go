package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

// Define a generic interface for your item
type Item interface{}

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Create a new DynamoDB session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new DynamoDB client
	db := dynamodb.New(sess)

	// Unmarshal the request body into an Item interface
	var item Item
	err = json.Unmarshal([]byte(event.Body), &item)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	// Convert the item to a map[string]*dynamodb.AttributeValue
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Get an input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("AWS_DYNAMO_TABLE_NAME")),
		Key:       av,
	}

	// Get the item in DynamoDB
	result, err := db.GetItem(input)
	if err != nil {
	    return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Unmarshal the DynamoDB result into an Item interface
	var receivedItem Item
	err = dynamodbattribute.UnmarshalMap(result.Item, &receivedItem)
	if err != nil {
	    return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Return the received item in the response body
	response := fmt.Sprintf("Item: %+v", receivedItem)
	return events.APIGatewayProxyResponse{
	    StatusCode: 200,
	    Body:       response,
	}, nil

}

func main() {
	lambda.Start(handler)
}
