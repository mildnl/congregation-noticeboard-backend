# Congregation Noticeboard Backend

This is the backend component of the Congregation Noticeboard application. It provides the functionality to store items in DynamoDB using AWS Lambda and API Gateway.

## Getting Started

These instructions will guide you on how to set up and deploy the backend component of the Congregation Noticeboard application.

### Prerequisites

- Go programming language (version 1.16 or later)
- AWS CLI configured with appropriate permissions
- AWS SDK for Go

### Installation

1. Clone the repository:

   ```shell
   git clone https://github.com/mildnl/congregation-noticeboard-backend.git
   ```
2. Navigate to the project directory:
```shell
cd congregation-noticeboard-backend
```
3. Install the required dependencies:
```shell
go mod download
```
4. Set up your AWS credentials using the AWS CLI:
```shell
aws configure
```
Make sure you have the necessary permissions to create and manage AWS Lambda functions, API Gateway, and DynamoDB tables.
### Deployment
To deploy the backend component, follow these steps:

1. Update the tableName constant in the main.go file to the desired DynamoDB table name.
2. Deploy the Lambda function using the AWS CLI:
```shell
aws lambda create-function --function-name congregation-noticeboard-backend \
    --runtime go1.x --zip-file fileb://main.zip --handler main \
    --role <your-iam-role-arn>
```
Replace `<your-iam-role-arn>` with the ARN of the IAM role that has the necessary permissions for the Lambda function.
3. Create an API Gateway REST API:
```shell
aws apigatewayv2 create-api --name congregation-noticeboard-api --protocol-type HTTP --target congregation-noticeboard-backend
```
4. Update the Lambda function configuration with the API Gateway integration:
```shell
aws lambda update-function-configuration --function-name congregation-noticeboard-backend --environment "Variables={API_ENDPOINT=<your-api-endpoint>}"
```
Replace `<your-api-endpoint>` with the URL of the API Gateway endpoint created in the previous step.
5. Test the backend by making HTTP requests to the API Gateway endpoint.
### Usage
The backend provides a single API endpoint for storing items in DynamoDB. Use an HTTP POST request to the API endpoint with the following payload:

```json
{
  "Id": 123,
  "name": "Test Item"
}
```
The `Id` field should be a unique integer, and the name field can be any string.

The response will indicate whether the item was stored successfully in DynamoDB.

### Contributing
Contributions are welcome! If you find any issues or would like to suggest improvements, please create a GitHub issue or submit a pull request.

### License
This project is licensed under the [MIT License](README.md).