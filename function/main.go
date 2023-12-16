package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var url = os.Getenv("get_url")

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s", url))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return events.APIGatewayProxyResponse{}, err
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return events.APIGatewayProxyResponse{}, err
	}

	err = os.WriteFile("/tmp/dinnerMenu.json", body, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello %v", body),
		StatusCode: 200,
	}, nil
}
