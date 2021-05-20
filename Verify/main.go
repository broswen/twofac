package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/broswen/twofac/code"
)

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

func ToJSON(o interface{}) string {
	j, _ := json.Marshal(o)
	return string(j)
}

var dynamoClient *dynamodb.Client

func Handler(ctx context.Context, event Request) (Response, error) {

	phoneNumber := event.PathParameters["number"]
	c := event.PathParameters["code"]

	getItemResponse, err := dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("CODETABLE")),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: phoneNumber},
		},
	})

	if err != nil {
		log.Println(err)
		return Response{StatusCode: http.StatusInternalServerError}, nil
	}

	if getItemResponse.Item == nil {
		log.Println("valid code not found for ", phoneNumber)
		return Response{StatusCode: http.StatusNotFound}, nil
	}

	prevCode := getItemResponse.Item["code"].(*types.AttributeValueMemberS).Value
	expires, _ := strconv.Atoi(getItemResponse.Item["expires"].(*types.AttributeValueMemberN).Value)
	status := getItemResponse.Item["status"].(*types.AttributeValueMemberS).Value

	if c != prevCode {
		return Response{StatusCode: http.StatusUnauthorized}, nil
	}

	if status == code.VERIFIED {
		return Response{StatusCode: http.StatusUnauthorized}, nil
	}

	if expires < int(time.Now().Unix()) {
		return Response{StatusCode: http.StatusUnauthorized}, nil
	}

	resp := Response{
		StatusCode: 200,
	}

	return resp, nil
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	lambda.Start(Handler)
}
