package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"

	util "github.com/mildnl/congregation-noticeboard-backend/util"
)

type TestItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (ti *TestItem) MarshalJSON() ([]byte, error) {
	type Alias TestItem
	return json.Marshal(&struct {
		*Alias
		CreatedAt int64 `json:"created_at"`
		UpdatedAt int64 `json:"updated_at"`
	}{
		Alias:     (*Alias)(ti),
		CreatedAt: ti.CreatedAt.Unix(),
		UpdatedAt: ti.UpdatedAt.Unix(),
	})
}

func (ti *TestItem) UnmarshalJSON(data []byte) error {
	type Alias TestItem
	aux := &struct {
		CreatedAt int64 `json:"created_at"`
		UpdatedAt int64 `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(ti),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	ti.CreatedAt = time.Unix(aux.CreatedAt, 0)
	ti.UpdatedAt = time.Unix(aux.UpdatedAt, 0)
	return nil
}

func TestHandler(t *testing.T) {
	// Create N sample items
	numItems := 3
	ids := make([]int, numItems)
	for i := 0; i < numItems; i++ {
		ids[i] = setup(t)
	}

	// Create a sample request with IDs
	requestBody, _ := json.Marshal(map[string][]int{
		"ids": ids,
	})
	request := events.APIGatewayProxyRequest{
		Body: string(requestBody),
	}

	// Call the handler function
	response, err := handler(context.Background(), request)

	// Check if there was an error
	assert.NoError(t, err)

	// Check the response status code
	assert.Equal(t, 200, response.StatusCode)

	// Parse the response body
	var receivedItems []TestItem
	err = json.Unmarshal([]byte(response.Body), &receivedItems)

	// Check if there was an error while parsing the response body
	assert.NoError(t, err)
	assert.Equal(t, numItems, len(receivedItems))
	for i := 0; i < numItems; i++ {
		assert.Equal(t, ids[i], receivedItems[i].ID)
		assert.Equal(t, "Test Title", receivedItems[i].Title)
		assert.Equal(t, "Test Content", receivedItems[i].Content)
		assert.Equal(t, "Gopher Test", receivedItems[i].Author)
		assert.Equal(t, time.Now().Unix(), receivedItems[i].CreatedAt.Unix())
		assert.Equal(t, time.Now().Unix(), receivedItems[i].UpdatedAt.Unix())
		assert.Nil(t, receivedItems[i].DeletedAt)
	}
	for i := 0; i < numItems; i++ {
		defer teardown(t, ids[i])
	}
}

func setup(t *testing.T) int {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	id := r.Intn(1000)

	testItem, _ := json.Marshal(map[string]interface{}{
		"Id": id,
		"Title": "Test Title",
		"Content": "Test Content",
		"Author": "Gopher Test",
		"CreatedAt": time.Now(),
		"UpdatedAt": time.Now(),
		"DeletedAt": nil,
		},
	)

	// Create the testing entry
	storedId, err := util.StoreItem(id, testItem)
	assert.NoError(t, err)
	assert.Equal(t, id, storedId)

	return id
}

func teardown(t *testing.T, id int) {
	// Delete the testing entry
	// Adjust this code based on your implementation
	err := util.DeleteItem(id)
	assert.NoError(t, err)

	// Verify the deletion
	// Adjust this code based on your implementation
	item, err := util.GetItem(id)
	assert.NoError(t, err)
	assert.Nil(t, item)
}
