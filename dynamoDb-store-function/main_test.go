package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"
	"github.com/mildnl/congregation-noticeboard-backend/dynamoDb-util"
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
	err := dynamoDb_util.DeleteItem(id)
	assert.NoError(t, err)

	// Verify the deletion
	item, err := dynamoDb_util.GetItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}

