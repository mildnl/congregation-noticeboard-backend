package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var tableName string

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	// Retrieve the environment variables
	tableName = os.Getenv("AWS_DYNAMO_TABLE_NAME")
}

func TestHandler(t *testing.T) {
	// Test setup
	id := setup(t)

	// Prepare a sample APIGatewayProxyRequest for testing
	requestBody := fmt.Sprintf(`{"Id": %d, "name": "Test Item"}`, id)
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := handler(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Assert the expected response body
	assert.Equal(t, fmt.Sprintf("Item stored successfully: map[Id:%d name:Test Item]", id), response.Body)

	// Test teardown
	teardown(t, id)
}

func setup(t *testing.T) int {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	id := r.Intn(1000)
	return id
}

func teardown(t *testing.T, id int) {
	// Delete the testing entry
	err := deleteItem(id)
	assert.NoError(t, err)

	// Verify the deletion
	item, err := getItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}

// deleteItem deletes the item with the specified ID from DynamoDB
func deleteItem(id int) error {
	// Create a new DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	svc := dynamodb.New(sess)

	// Create the input for the DeleteItem operation
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				N: aws.String(strconv.Itoa(id)),
			},
		},
	}

	// Perform the DeleteItem operation
	_, err := svc.DeleteItem(input)
	if err != nil {
		return err
	}

	return nil
}

// getItem retrieves the item with the specified ID from DynamoDB
func getItem(id int) (*Item, error) {
	// Create a new DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	svc := dynamodb.New(sess)

	// Create the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				N: aws.String(strconv.Itoa(id)),
			},
		},
	}

	// Perform the GetItem operation
	result, err := svc.GetItem(input)
	if err != nil {
		return nil, err
	}

	// Check if the item exists
	if len(result.Item) == 0 {
		return nil, nil
	}

	// Unmarshal the item into a struct
	var item *Item
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}