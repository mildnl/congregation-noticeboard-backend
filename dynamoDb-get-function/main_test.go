package main

import (
	"context"
	"encoding/json"
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
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	id := r.Intn(1000)

	// Prepare a sample item and store it in DynamoDB
	err := storeItem(id, "Test Item")
	assert.NoError(t, err)

	// Prepare a sample APIGatewayProxyRequest for testing
	request := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"id": strconv.Itoa(id),
		},
	}

	// Invoke the handler function
	response, err := handler(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Unmarshal the response body
	var item Item
	err = json.Unmarshal([]byte(response.Body), &item)
	assert.NoError(t, err)

	// Verify the item's ID and name
	assert.Equal(t, id, item.Id)
	assert.Equal(t, "Test Item", item.Name)

	// Delete the testing item
	err = deleteItem(id)
	assert.NoError(t, err)

	// Verify the deletion
	item, err = getItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}

// storeItem stores the item with the specified ID and name in DynamoDB
func storeItem(id int, name string) error {
	// Create a new DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	svc := dynamodb.New(sess)

	// Create the item
	item := Item{
		Id:   id,
		Name: name,
	}

	// Marshal the item into a map[string]*dynamodb.AttributeValue
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Perform the PutItem operation
	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}

	return nil
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
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return item, nil
}
