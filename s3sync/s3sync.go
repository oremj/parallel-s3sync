package s3sync

import (
	"bytes"
	"errors"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

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

func New(awsConfig *aws.Config) *S3Sync {
	return &S3Sync{
		AWSConfig: awsConfig,
	}
}

func (s *S3Sync) Sync(source, target string, workers int) error {
	if isLocalPath(source) && isS3Path(target) {
		s3url, err := parseS3Path(target)
		if err != nil {
			return err
		}

		return s.syncLocalToS3(source, s3url.Host, s3url.Path, workers)
	}

	return errors.New("Operation not supported")
}

type S3Sync struct {
	AWSConfig    *aws.Config
	CopySymlinks bool
}

func cleanS3Path(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func (s *S3Sync) syncLocalToS3(source, bucket, prefix string, workers int) error {
	prefix = cleanS3Path(prefix)

	bucketIndex, err := s.bucketIndex(bucket, prefix)
	if err != nil {
		return err
	}

	fileChan := make(chan *localToS3Input, workers*1000)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			s3Svc := s3.New(s.AWSConfig)
			for f := range fileChan {
				start := time.Now()
				log.Println("START:", f.LocalPath, f.Params.Key)

				err := localToS3(s3Svc, f)
				if err != nil {
					log.Print(err)
				}

				log.Printf("DONE (%s): %s %s", time.Since(start), f.LocalPath, f.Params.Key)
			}
		}()
	}

	fileFilter := func(path string, info os.FileInfo) bool {
		if s.CopySymlinks && info.Mode()&os.ModeSymlink != 0 {
			return true
		}
		if info.Mode().IsRegular() {
			return true
		}

		return false
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
		}
		if !fileFilter(path, info) {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		key := prefix + relPath

		if bucketIndex.Exists(key, info.Size()) {
			log.Println("Exists:", key)
			return nil
		}

		params := &s3.PutObjectInput{
			Key:         aws.String(key),
			Bucket:      aws.String(bucket),
			ContentType: aws.String(ContentType(key)),
		}

		fileChan <- &localToS3Input{LocalPath: path, Params: params, Info: info}
		return nil
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

func (s *S3Sync) bucketIndex(bucket, prefix string) (S3KeyMap, error) {
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

type localToS3Input struct {
	LocalPath string
	Params    *s3.PutObjectInput
	Info      os.FileInfo
}

func localToS3(s3Svc *s3.S3, in *localToS3Input) error {
	metadata := make(map[string]*string)
	if in.Info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(in.LocalPath)
		if err != nil {
			return err
		}
		metadata["mode"] = aws.String("S_IFLNK")
		in.Params.Body = bytes.NewReader([]byte(target))

	} else if in.Info.Mode().IsRegular() {
		file, err := os.Open(in.LocalPath)
		if err != nil {
			return err
		}
		defer file.Close()

		in.Params.Body = file
	}

	if len(metadata) > 0 {
		in.Params.Metadata = &metadata
	}

	_, err := s3Svc.PutObject(in.Params)
	return err
}
