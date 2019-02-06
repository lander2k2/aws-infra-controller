package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Bucket struct {
	Region string
	Name   string
}

type Object struct {
	Region   string
	Location string
	Body     string
}

func (bucket *Bucket) Create() error {
	svc := s3.New(session.New(&aws.Config{Region: aws.String(bucket.Region)}))

	if _, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (bucket *Bucket) Describe() error {
	return nil
}

func (bucket *Bucket) Delete() error {
	svc := s3.New(session.New(&aws.Config{Region: aws.String(bucket.Region)}))

	if _, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket.Name),
	}); err != nil {
		return err
	}

	return nil
}

func (object *Object) Put() error {
	svc := s3.New(session.New(&aws.Config{Region: aws.String(object.Region)}))

	if _, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(object.Location),
		Body:   aws.ReadSeekCloser(strings.NewReader(object.Body)),
		Key:    aws.String("join"),
	}); err != nil {
		return err
	}

	return nil
}

func (object *Object) Get() error {
	return nil
}
