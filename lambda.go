package main

import (
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"encoding/base64"
	"fmt"
)

type Request struct {}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func Handler(request Request) (Response, error) {
	parseFlags()
	run()
	return Response {
		Message: fmt.Sprintf("Processed request"),
		Ok:      true,
	}, nil
}

func kmsDecrypt(bas64EncodedSecret string) (string, error) {
	secretBytes, _ := base64.StdEncoding.DecodeString(bas64EncodedSecret)
	kmsClient := kms.New(session.New(&aws.Config{
		Region: aws.String("eu-central-1"),
	}))

	params := &kms.DecryptInput{
		CiphertextBlob: secretBytes,
	}

	resp, err := kmsClient.Decrypt(params)
	if err != nil {
		return "", err
	}
	return string(resp.Plaintext), nil
}

func parseLambdaFlags(args string) {
	if strings.Contains(args, "-cloudflare") {
		useCloudflare = true
	}
	if strings.Contains(args, "-add-missing") {
		addMissing = true
	}
	if strings.Contains(args, "-delete-obsolete") {
		deleteObsolete = true
	}
	if strings.Contains(args, "-acknowledge") {
		acknowledge = true
	}
	if strings.Contains(args, "-list-ss-ids") {
		listSSIDs = true
	}
	if strings.Contains(args, "-version") {
		printVersion = true
	}
}
