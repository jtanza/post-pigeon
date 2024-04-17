package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const PostBucket = "dev-postpigeon-posts"

func CreateS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return s3.NewFromConfig(cfg)
}

func UploadPost(client *s3.Client, postUUID string, html string) (string, error) {
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(PostBucket),
		Key:    aws.String(postUUID),
		Body:   strings.NewReader(html),
	})
	if err != nil {
		return "", err
	}
	return toS3Url(PostBucket, postUUID), nil
}

func toS3Url(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}
