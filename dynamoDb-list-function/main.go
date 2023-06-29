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

	// Unmarshal the request body into a struct containing the IDs
	var request struct {
		Ids []int `json:"ids"`
	}
	err = json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	// Create a slice to hold the BatchGetItem request items
	var batchGetItems []*dynamodb.KeysAndAttributes

	// Iterate over the IDs and create BatchGetItem request items for each ID
	for _, id := range request.Ids {
		// Convert the ID to a DynamoDB attribute value
		idAttributeValue, err := dynamodbattribute.Marshal(id)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}

		// Create a BatchGetItem request item for the ID
		item := map[string]*dynamodb.AttributeValue{
			"id": idAttributeValue,
		}

		// Create a KeysAndAttributes struct for the request item
		keysAndAttributes := &dynamodb.KeysAndAttributes{
			Keys: []map[string]*dynamodb.AttributeValue{item},
		}

		// Add the KeysAndAttributes struct to the slice
		batchGetItems = append(batchGetItems, keysAndAttributes)
	}

	// Create the BatchGetItem input
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			os.Getenv("AWS_DYNAMO_TABLE_NAME"): {
				Keys:             nil,
				ConsistentRead:   aws.Bool(true), // Set to true if you want consistent reads
			},
		},
	}

	// Set the batchGetItems slice as the value of the table in the input
	keys := make([]map[string]*dynamodb.AttributeValue, len(batchGetItems))
	for i, keysAndAttributes := range batchGetItems {
		keys[i] = keysAndAttributes.Keys[0]
	}
	input.RequestItems[os.Getenv("AWS_DYNAMO_TABLE_NAME")].Keys = keys
	

	// Get the items from DynamoDB
	result, err := db.BatchGetItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Unmarshal the DynamoDB result into a slice of Item interface
	var receivedItems []Item
	err = dynamodbattribute.UnmarshalListOfMaps(result.Responses[os.Getenv("AWS_DYNAMO_TABLE_NAME")], &receivedItems)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Return the received items in the response body
	response := fmt.Sprintf("Received Items: %+v", receivedItems)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       response,
	}, nil
}

func main() {
	lambda.Start(handler)
}
