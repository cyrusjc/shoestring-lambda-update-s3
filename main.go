package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const Bucket = "shoestring-cafe-react"
const FileName = "dinnerMenu.json"

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	chanBlock := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	go getSheetsDataFunc(chanBlock, &wg)
	go uploadToS3(chanBlock, &wg)

	wg.Wait()

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Success!"),
		StatusCode: 200,
	}, nil
}

func getSheetsDataFunc(chanBlock chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	var url = os.Getenv("get_url")
	resp, err := http.Get(fmt.Sprintf("%s", url))
	if err != nil {
		log.Println("Error GET request: ", err)
		os.Exit(0)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(0)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		os.Exit(0)
	}

	err = os.WriteFile("/tmp/dinnerMenu.json", body, 0666)
	chanBlock <- true
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	return
}

func uploadToS3(chanBlock chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	uploader := s3manager.NewUploader(sess)

	<-chanBlock
	file, err := os.Open("/tmp/" + FileName)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
		os.Exit(0)
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(FileName),
		Body:   file,
	})

	if err != nil {
		exitErrorf("Unable to upload %q to %q, %v", FileName, Bucket, err)
		os.Exit(0)
	}

}
