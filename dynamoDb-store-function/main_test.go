package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	util "github.com/mildnl/congregation-noticeboard-backend/util"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	// Test setup
	id := setup(t)

	// Prepare a sample APIGatewayProxyRequest for testing
	requestBody := fmt.Sprintf(`{"Id": %d, "name": "Test Item"}`, id)
	request := events.APIGatewayProxyRequest{
		Body: requestBody,
	}

	// Invoke the handler function
	response, err := Handler(context.Background(), request)
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
	err := util.DeleteItem(id)
	assert.NoError(t, err)

	// Verify the deletion
	item, err := util.GetItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}

