package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	PostBucket = "dev-postpigeon-posts"
	maxTitle   = 50
)

func CreateS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return s3.NewFromConfig(cfg)
}

func UploadPost(client *s3.Client, html string) (string, error) {
	key := genKey()
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(PostBucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(html),
	})
	if err != nil {
		return "", err
	}
	return toS3Url(PostBucket, key), nil
}

func toS3Url(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}

func genKey() string {
	// title := strings.ToLower(strings.ReplaceAll(request.Title, " ", "-"))
	// if len(title) > maxTitle {
	// 	title = title[:maxTitle]
	// }

	return fmt.Sprintf(
		"%s-%s-%d",
		"request.Author",
		"title",
		time.Now().Unix(),
	)
}
