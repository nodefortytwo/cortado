package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	// Initialize a session in eu-west-1 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)

	// Create S3 service client
	svc := s3.New(sess)

	bucket := os.Args[1]
	prefix := os.Args[2]

	// Get the list of items
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(prefix)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	key := *resp.Contents[0].Key

	fpath := os.TempDir() + "cortado"
	downloadFile(bucket, key, fpath, sess)
	editFile(fpath)
	uploadFile(bucket, key, fpath, sess)

}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func editFile(fpath string) {
	cmd := exec.Command("vim", fpath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Printf("2")
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error while editing. Error: %v\n", err)
	} else {
		log.Printf("Successfully edited.")
	}
}

func downloadFile(bucket string, key string, fpath string, sess client.ConfigProvider) {
	file, err := os.Create(fpath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		exitErrorf("Unable to download item %q, %v", key, err)
	}

	file.Close()
}

func uploadFile(bucket string, key string, fpath string, sess client.ConfigProvider) {
	file, err := os.Open(fpath)

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		exitErrorf("Unable to upload %q to %q, %v", key, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", key, bucket)
}
