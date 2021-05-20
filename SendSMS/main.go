package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/broswen/twofac/code"
)

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

func ToJSON(o interface{}) string {
	j, _ := json.Marshal(o)
	return string(j)
}

var dynamoClient *dynamodb.Client
var snsClient *sns.Client

func Handler(ctx context.Context, event Request) (Response, error) {

	phoneNumber := event.PathParameters["number"]
	xapikey := event.Headers["x-api-key"]

	c := code.Generate(6)
	log.Println("generated code", c)

	_, err := dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("CODETABLE")),
		Item: map[string]types.AttributeValue{
			"PK":      &types.AttributeValueMemberS{Value: phoneNumber},
			"code":    &types.AttributeValueMemberS{Value: c},
			"status":  &types.AttributeValueMemberS{Value: code.PENDING},
			"expires": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Add(time.Minute*5).Unix())},
		},
	})

	if err != nil {
		log.Println(err)
		return Response{StatusCode: http.StatusInternalServerError}, nil
	}

	publishResponse, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		Message:     aws.String(c),
		PhoneNumber: aws.String(phoneNumber),
	})
	if err != nil {
		log.Println(err)
		return Response{StatusCode: http.StatusInternalServerError}, nil
	}

	log.Printf("(%s) sent code to phone number %s", *publishResponse.MessageId, phoneNumber)

	msg := struct {
		Xapikey     string
		Code        string
		PhoneNumber string
	}{
		xapikey,
		c,
		phoneNumber,
	}

	log.Println(ToJSON(msg))

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
	snsClient = sns.NewFromConfig(cfg)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	lambda.Start(Handler)
}
