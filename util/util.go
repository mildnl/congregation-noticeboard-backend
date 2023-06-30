package util

import (
	cryptRand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"

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
		return 0, fmt.Errorf("failed to store item with ID %d: %w", id, err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Perform the PutItem operation to store the item in DynamoDB
	_, err = svc.PutItem(input)
	if err != nil {
		return 0, fmt.Errorf("failed to store item with ID %d: %w", id, err)
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

// generatePassword generates a password that satisfies the Cognito password policy requirements.
func GeneratePassword() string {
	// Generate the password
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Define the password policy requirements
	minLength := 8
	hasDigit := true
	hasLower := true
	hasUpper := true
	hasSpecial := true
	specialChars := "!@#$%^&*()-_=+{}[]|\\;:'\"<>,.?/~`"

	// Define the character sets
	digits := "0123456789"
	lowerChars := "abcdefghijklmnopqrstuvwxyz"
	upperChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet := specialChars

	// Initialize the password
	password := ""

	// Add at least one character from each required character set
	if hasDigit {
		password += string(digits[random.Intn(len(digits))])
	}
	if hasLower {
		password += string(lowerChars[random.Intn(len(lowerChars))])
	}
	if hasUpper {
		password += string(upperChars[random.Intn(len(upperChars))])
	}
	if hasSpecial {
		password += string(specialCharSet[random.Intn(len(specialCharSet))])
	}

	// Generate the remaining characters
	for i := len(password); i < minLength; i++ {
		charSet := ""
		if hasDigit {
			charSet += digits
		}
		if hasLower {
			charSet += lowerChars
		}
		if hasUpper {
			charSet += upperChars
		}
		if hasSpecial {
			charSet += specialCharSet
		}

		password += string(charSet[random.Intn(len(charSet))])
	}

	// Shuffle the password string
	shuffled := []rune(password)
	random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	
	return password
}

// GenerateAccessToken generates a random access token.
func GenerateAccessToken() (string, error) {
	tokenBytes := make([]byte, 32) // Generate a 256-bit random token
	_, err := cryptRand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	accessToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Remove characters that don't match the required pattern
	re := regexp.MustCompile("[^A-Za-z0-9-_=.]")
	accessToken = re.ReplaceAllString(accessToken, "")

	return accessToken, nil
}