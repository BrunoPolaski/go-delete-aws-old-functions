package main

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("region"),
	}))
	svc := lambda.New(sess)

	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String("lambda-function"),
	}
	result, _ := svc.ListVersionsByFunction(input)

	var wg sync.WaitGroup

	for _, version := range result.Versions {
		if *version.Version != "$LATEST" {
			wg.Add(1)

			go func(version string) {
				defer wg.Done()
				fmt.Println("Deleting version:", version)
				delInput := &lambda.DeleteFunctionInput{
					FunctionName: aws.String("lambda-function"),
					Qualifier:    aws.String(version),
				}
				_, delErr := svc.DeleteFunction(delInput)
				if delErr != nil {
					fmt.Println("Error deleting version", version, ":", delErr)
				} else {
					fmt.Println("Deleted version", version)
				}
			}(*version.Version)
		}
	}

	wg.Wait()
	fmt.Println("All versions processed.")
}
