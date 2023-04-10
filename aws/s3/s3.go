package s3

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/ripemd160"
)

// Manager contains s3 client functionalities
type Manager struct {
	bucketName string
	region     string
}

// NewManager initiates a new s3 manager
func NewManager(bucket string, region string) *Manager {
	return &Manager{
		bucketName: bucket,
		region:     region,
	}
}

// Save data as a object in s3
func (m *Manager) Save(payload []byte) error {

	var bucket, key string
	var timeout time.Duration

	// generating hash and create object name
	md := ripemd160.New()
	keyName, err := io.WriteString(md, string(payload[:]))

	objName := fmt.Sprintf("%v-%x", time.Now().Unix(), keyName)
	b := flag.Lookup("b")
	k := flag.Lookup("k")
	d := flag.Lookup("d")

	if b == nil && k == nil && d == nil {
		flag.StringVar(&bucket, "b", m.bucketName, "Bucket name.")
		flag.StringVar(&key, "k", objName, "Object key name.")
		flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
		flag.Parse()
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))
	svc := s3.New(sess)

	r := bytes.NewReader(payload)

	// Uploads the object to S3. The Context will interrupt the request if the
	// timeout expires.
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(objName),
		Body:   r,
	})
	return err
}

// GetKeys get all s3 bucket objects keys
func (m *Manager) GetKeys() ([]string, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)

	var objects []string
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(m.bucketName),
	}, func(p *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range p.Contents {
			objects = append(objects, aws.StringValue(o.Key))
		}
		return true // continue paging
	})
	if err != nil {
		return nil, err
	}

	return objects, nil
}

// Get get a single s3 object by key
func (m *Manager) Get(key string) ([]byte, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(obj.Body)
	return body, nil
}

// Delete delete s3 object by key
func (m *Manager) Delete(key string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))

	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

// SaveWithKey save data with specific key/path as an object in s3
func (m *Manager) SaveWithKey(payload []byte, objectKey string) error {
	var bucket, key string
	var timeout time.Duration

	b := flag.Lookup("b")
	k := flag.Lookup("k")
	d := flag.Lookup("d")

	if b == nil && k == nil && d == nil {
		flag.StringVar(&bucket, "b", m.bucketName, "Bucket name.")
		flag.StringVar(&key, "k", objectKey, "Object key name.")
		flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
		flag.Parse()
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))
	svc := s3.New(sess)

	r := bytes.NewReader(payload)

	// Uploads the object to S3. The Context will interrupt the request if the
	// timeout expires.
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(m.bucketName),
		Key:    aws.String(objectKey),
		Body:   r,
	})
	return err
}

// GetFilesFromSubFolder returns files from a subfolder in the bucket
func (m *Manager) GetFilesFromSubFolder(folder string) ([]string, error) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))
	svc := s3.New(sess)

	var objects []string
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(m.bucketName),
		Prefix: aws.String(folder),
	}, func(p *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range p.Contents {
			objects = append(objects, aws.StringValue(o.Key))
		}
		return true // continue paging
	})
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (m *Manager) UploadMultipart(fileName string, path string) error {
	const (
		maxPartSize = int64(50 * 1024 * 1024)
		maxRetries  = 3
	)
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(m.region)}))
	svc := s3.New(sess)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("err opening file: %s", err)
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	fileType := http.DetectContentType(buffer)
	file.Read(buffer)

	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(m.bucketName),
		Key:         aws.String(path),
		ContentType: aws.String(fileType),
	}

	resp, err := svc.CreateMultipartUpload(input)
	if err != nil {
		return err
	}

	var curr, partLength int64
	var remaining = size
	var completedParts []*s3.CompletedPart
	partNumber := 1
	for curr = 0; remaining != 0; curr += partLength {
		if remaining < maxPartSize {
			partLength = remaining
		} else {
			partLength = maxPartSize
		}
		completedPart, err := uploadPart(svc, resp, file, partNumber, maxRetries)
		if err != nil {
			fmt.Println(err.Error())
			err := abortMultipartUpload(svc, resp)
			if err != nil {
				fmt.Println(err.Error())
			}
			return err
		}
		remaining -= partLength
		partNumber++
		completedParts = append(completedParts, completedPart)
	}

	_, err = completeMultipartUpload(svc, resp, completedParts)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully uploaded file")

	return nil
}

func completeMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}
	return svc.CompleteMultipartUpload(completeInput)
}

func uploadPart(svc *s3.S3, resp *s3.CreateMultipartUploadOutput, file *os.File, partNumber int, maxRetries int) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:       file,
		Bucket:     resp.Bucket,
		Key:        resp.Key,
		PartNumber: aws.Int64(int64(partNumber)),
		UploadId:   resp.UploadId,
	}

	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPart(partInput)
		if err != nil {
			if tryNum == maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			fmt.Printf("Retrying to upload part #%v\n", partNumber)
			tryNum++
		} else {
			fmt.Printf("Uploaded part #%v\n", partNumber)
			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}
	return nil, nil
}

func abortMultipartUpload(svc *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	fmt.Println("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := svc.AbortMultipartUpload(abortInput)
	return err
}
