package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mildnl/congregation-noticeboard-backend/dynamoDb-util"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)


// Mock DynamoDB client for testing
type mockDynamoDBClient struct{}

func (m *mockDynamoDBClient) PutItemWithContext(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Simulate a successful PutItem operation
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Item stored successfully",
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func TestHandler(t *testing.T) {
	// Test setup
	id := setup(t)

	// Prepare a sample APIGatewayProxyRequest for testing
	requestBody := fmt.Sprintf(`{ "Id": %d }`, id)
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := handler(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Assert the expected response body
	assert.Equal(t, fmt.Sprintf("Item deleted successfully: map[Id:%d]", id), response.Body)

	// Retrieve the item using the getItem function
	item, err := dynamoDb_util.GetItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}



func setup(t *testing.T) int {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	id := r.Intn(1000)

	testItem := map[string]interface{}{
		"Id":   id,
		"name": "Test Item",
	}
	// Store the testing entry
	storedId, err := dynamoDb_util.StoreItem(id, testItem )
	if err != nil {
		t.Errorf("Error storing item: %s", err)
		return 0
	}
	if storedId != id {
		t.Errorf("Expected id %d, got %d", id, storedId)
		return 0
	}
	return id
}