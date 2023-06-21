package dynamoDb_util

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/joho/godotenv"
)

var tableName string

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	// Retrieve the environment variable for table name
	tableName = os.Getenv("AWS_DYNAMO_TABLE_NAME")
}

// StoreItem stores an item in DynamoDB with the given ID
func StoreItem(id int, item interface{}) (int, error) {
	// Create a new DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	svc := dynamodb.New(sess)

	// Marshal the item into a DynamoDB attribute value map
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return 0, err
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Perform the PutItem operation to store the item in DynamoDB
	_, err = svc.PutItem(input)
	if err != nil {
		return 0, err
	}

	// Return the stored ID
	return id, nil
}

// GetItem retrieves the item with the specified ID from DynamoDB
func GetItem(id int) (interface{}, error) {
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

	// Unmarshal the item into a map
	item := make(map[string]interface{})
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteItem deletes the item with the specified ID from DynamoDB
func DeleteItem(id int) error {
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
