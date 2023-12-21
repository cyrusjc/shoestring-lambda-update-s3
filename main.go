package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const Bucket = "shoestring-cafe-react"

type Menu struct {
	name string
	url  string
}

func newMenu(name string, url string) *Menu {
	return &Menu{
		name: name,
		url:  url,
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	dinner := newMenu("dinner", fmt.Sprintf("%v", os.Getenv("get_url")))
	lunch := newMenu("lunch", fmt.Sprintf("%v", os.Getenv("get_url_lunch")))
	opentable := newMenu("opentable", fmt.Sprintf("%v", os.Getenv("get_url_opentable")))

	menus := []*Menu{dinner, lunch, opentable}

	var wg sync.WaitGroup
	for _, menu := range menus {
		chanBlock := make(chan bool)
		wg.Add(2)
		go getSheetsDataFunc(chanBlock, &wg, menu)
		go uploadToS3(chanBlock, &wg, menu)
	}

	wg.Wait()

	err := invalidateCloudfront()
	if err != nil {
		log.Println("Error GET request: ", err)
		os.Exit(0)
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Success!"),
		StatusCode: 200,
	}, nil
}

func getSheetsDataFunc(chanBlock chan bool, wg *sync.WaitGroup, menu *Menu) {

	FileName := menu.name + "Menu.json"

	defer wg.Done()

	resp, err := http.Get(menu.url)
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

	err = os.WriteFile("/tmp/"+FileName, body, 0666)
	chanBlock <- true
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	return
}

func uploadToS3(chanBlock chan bool, wg *sync.WaitGroup, menu *Menu) {

	FileName := menu.name + "Menu.json"
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
	}

}

func invalidateCloudfront() error {
	now := time.Now()
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	// Example sending a request using the CreateInvalidationRequest method.
	svc := cloudfront.New(sess)

	params := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(fmt.Sprintf("%v", os.Getenv("DIST_ID"))),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(
				fmt.Sprintf("goinvali%s", now.Format("2006/01/02,15:04:05"))),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(1),
				Items: []*string{
					aws.String("/*"),
				},
			},
		},
	}

	req, resp := svc.CreateInvalidationRequest(params)
	req.Send()

	fmt.Println(resp)

	return nil

}
