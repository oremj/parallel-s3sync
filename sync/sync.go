package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

type Sync struct {
	AWSConfig *aws.Config
	Bucket    string
}

type S3KeyMap map[string]*s3.Object

func (s S3KeyMap) Exists(key string, size int64) bool {
	v, ok := s[key]
	if !ok {
		return false
	}

	return *v.Size == size
}

func (s *Sync) KeyIndex(prefix string) (S3KeyMap, error) {
	s3Svc := s3.New(s.AWSConfig)

	keymap := make(map[string]*s3.Object)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	}
	more_keys := true

	for more_keys {
		resp, err := s3Svc.ListObjects(params)

		if awserr := aws.Error(err); awserr != nil {
			fmt.Println("Error:", awserr.Code, awserr.Message)
			continue
		} else if err != nil {
			panic(err)
		}

		more_keys = *resp.IsTruncated
		var o *s3.Object
		for _, o = range resp.Contents {
			keymap[*o.Key] = o
		}
		if o != nil {
			params.Marker = aws.String(*o.Key)
		}
	}

	return keymap, nil
}

func (s *Sync) PutFile(localPath, key string) error {
	s3Svc := s3.New(s.AWSConfig)

	file, err := os.Open(localPath)

	if err != nil {
		return err
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	_, err = s3Svc.PutObject(params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sync) Sync(localPath, remotePath string) error {
	remotePath = strings.TrimPrefix(remotePath, "/")
	if remotePath != "" && !strings.HasSuffix(remotePath, "/") {
		remotePath += "/"
	}
	keyIndex, err := s.KeyIndex(remotePath)
	if err != nil {
		return err
	}

	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		relPath, err := filepath.Rel(localPath, path)
		key := remotePath + relPath

		if keyIndex.Exists(key, info.Size()) {
			fmt.Println("Exists:", key)
			return nil
		}

		err = s.PutFile(path, key)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}
