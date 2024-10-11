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
		// Credentials: credentials.NewSharedCredentials("", "default"), // Uncomment if you are not using default profile picked up from ~/.aws/credentials
	}))
	svc := lambda.New(sess)

	var wg sync.WaitGroup

	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String("lambda-function"),
	}

	var result *lambda.ListVersionsByFunctionOutput

	for {
		versions, err := svc.ListVersionsByFunction(input)
		if err != nil {
			fmt.Println("Error listing versions:", err)
			return
		}

		if result == nil {
			result = versions
		} else {
			result.Versions = append(result.Versions, versions.Versions...)
		}

		if versions.NextMarker == nil {
			break
		}

		input.Marker = versions.NextMarker
	}

	result.Versions = result.Versions[1:]
	result.Versions = result.Versions[:len(result.Versions)-1]

	for _, version := range result.Versions {
		fmt.Println("Processing version: ", *version.Version)
		wg.Add(1)
		go func(version string) {
			defer wg.Done()
			delInput := &lambda.DeleteFunctionInput{
				FunctionName: aws.String("lambda-function"),
				Qualifier:    aws.String(version),
			}
			_, delErr := svc.DeleteFunction(delInput)
			if delErr != nil {
				fmt.Println("Error deleting version ", version, ":", delErr)
			} else {
				fmt.Println("Deleted version: ", version)
			}
		}(*version.Version)
	}

	wg.Wait()
	fmt.Println("All versions processed.")
}
