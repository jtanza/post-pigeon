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

const postBucket = "post-pigeon.com"
const postPrefix = "posts"

type S3Client struct {
	client *s3.Client
}

func NewS3Client() S3Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return S3Client{s3.NewFromConfig(cfg)}
}

func (s S3Client) UploadPostObject(postUUID string, html string) (string, error) {
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(postBucket),
		Key:         aws.String(FormatKeyPath(postUUID)),
		Body:        strings.NewReader(html),
		ContentType: aws.String("text/html"),
	})
	if err != nil {
		return "", err
	}

	return toS3Url(postBucket, postUUID), nil
}

func (s S3Client) DeletePostObject(postUUID string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(postBucket),
		Key:    aws.String(FormatKeyPath(postUUID)),
	})
	return err
}

func FormatKeyPath(postUUID string) string {
	return fmt.Sprintf("%s/%s", postPrefix, postUUID)
}

func toS3Url(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}
