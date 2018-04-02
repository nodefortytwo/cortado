package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/c-bata/go-prompt"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n%s [options] {bucket}\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	region := flag.String("region", "eu-west-1", "AWS Region")
	prefix := flag.String("prefix", "", "Key prefix to use")
	editor := flag.String("editor", "vim", "Which editor to use, only a cli editor will function properly")
	flag.Parse()

	bucket := flag.Arg(0)
	if bucket == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize a session in eu-west-1 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*region)},
	)

	if err != nil {
		exitErrorf("Unable load config")
	}

	key := selectKey(bucket, *prefix, sess)

	if len(key) == 0 {
		exitErrorf("No key found, or selected")
	}

	fpath := os.TempDir() + randString(10)

	downloadFile(bucket, key, fpath, sess)

	defer cleanUp(fpath)

	hash, _ := md5sum(fpath)
	editFile(fpath, *editor)
	editedHash, _ := md5sum(fpath)

	if hash != editedHash {
		uploadFile(bucket, key, fpath, sess)
	} else {
		fmt.Printf("No changes made\n")
	}

}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func editFile(fpath string, editor string) {
	cmd := exec.Command(editor, fpath)

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

	fmt.Printf("Downloaded %q to %q\n", key, fpath)

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

func selectKey(bucket string, prefix string, sess client.ConfigProvider) string {
	// Create S3 service client
	svc := s3.New(sess)
	// Get the list of items
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(prefix)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	var key string

	if len(resp.Contents) == 0 {
		key = ""
	} else if len(resp.Contents) > 1 {
		fmt.Println("Please start typing key")
		key = prompt.Input("> ", buildCompleter(resp.Contents))
	} else {
		key = *resp.Contents[0].Key
	}

	return key
}

func randString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func buildCompleter(contents []*s3.Object) func(d prompt.Document) []prompt.Suggest {

	return func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{}

		for _, obj := range contents {
			s = append(s, prompt.Suggest{Text: *obj.Key})
		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}

func md5sum(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return
}

func cleanUp(fpath string) {
	fmt.Printf("Deleting %q \n", fpath)
	os.Remove(fpath)
}
