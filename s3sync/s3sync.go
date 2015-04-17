package s3sync

import (
	"errors"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

type syncFile struct {
	Key       string
	LocalPath string
	Info      os.FileInfo
}

func isLocalPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func isS3Path(path string) bool {
	return strings.HasPrefix(path, "s3://")
}

func parseS3Path(path string) (*url.URL, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "s3" || u.Host == "" || u.Path == "" {
		return nil, errors.New("Not a valid s3_path. Example: s3://bucket/path")
	}

	return u, nil
}

func Sync(awsConfig *aws.Config, source, target string, workers int) error {
	s := &s3sync{
		AWSConfig: awsConfig,
	}
	if isLocalPath(source) && isS3Path(target) {
		s3url, err := parseS3Path(target)
		if err != nil {
			return err
		}

		return s.syncLocalToS3(source, s3url.Host, s3url.Path, workers)
	}

	return errors.New("Operation not supported")
}

type s3sync struct {
	AWSConfig *aws.Config
}

func cleanS3Path(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func (s *s3sync) syncLocalToS3(source, bucket, prefix string, workers int) error {
	prefix = cleanS3Path(prefix)

	bucketIndex, err := s.bucketIndex(bucket, prefix)
	if err != nil {
		return err
	}

	fileChan := make(chan *syncFile, workers*1000)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			s3Svc := s3.New(s.AWSConfig)
			for f := range fileChan {
				log.Println("Putting:", f.LocalPath, f.Key)

				params := &s3.PutObjectInput{
					Key:         aws.String(f.Key),
					Bucket:      aws.String(bucket),
					ContentType: aws.String(ContentType(f.Key)),
				}
				err := s.localToS3(s3Svc, f.LocalPath, params)
				if err != nil {
					log.Print(err)
				}
			}
			wg.Done()
		}()
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return err
		}
		relPath, err := filepath.Rel(source, path)
		key := prefix + relPath

		if bucketIndex.Exists(key, info.Size()) {
			log.Println("Exists:", key)
			return nil
		}

		fileChan <- &syncFile{Key: key, LocalPath: path, Info: info}
		return err
	})
	close(fileChan)

	if err != nil {
		log.Print(err)
	}

	wg.Wait()

	return nil
}

type S3KeyMap map[string]*s3.Object

func (s S3KeyMap) Exists(key string, size int64) bool {
	v, ok := s[key]
	if !ok {
		return false
	}

	return *v.Size == size
}

func (s *s3sync) bucketIndex(bucket, prefix string) (S3KeyMap, error) {
	s3Svc := s3.New(s.AWSConfig)

	keymap := make(S3KeyMap)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	more_keys := true
	retries := 5
	for more_keys {
		resp, err := s3Svc.ListObjects(params)

		if awserr := aws.Error(err); awserr != nil {
			log.Println("Error:", awserr.Code, awserr.Message)
			retries--
			if retries < 0 {
				return nil, err
			}
			continue
		} else if err != nil {
			return nil, err
		}

		retries = 5

		more_keys = *resp.IsTruncated
		var o *s3.Object
		for _, o = range resp.Contents {
			keymap[*o.Key] = o
		}
		if o != nil {
			params.Marker = aws.String(*o.Key)
		} else {
			break
		}
	}

	return keymap, nil
}

func ContentType(path string) string {
	if tmp := mime.TypeByExtension(filepath.Ext(path)); tmp != "" {
		return tmp
	}
	return "binary/octet-stream"
}

func (s *s3sync) localToS3(s3Svc *s3.S3, localPath string, params *s3.PutObjectInput) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	params.Body = file

	_, err = s3Svc.PutObject(params)
	return err
}
