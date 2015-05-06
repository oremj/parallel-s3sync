package s3sync

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/oremj/parallel-s3sync/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/aws"
	"github.com/oremj/parallel-s3sync/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/service/s3"
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
	AWSConfig       *aws.Config
	CopySymlinks    bool
	ExcludePatterns []string
}

func cleanS3Path(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func (s *S3Sync) excludeFile(path string) bool {
	for _, pattern := range s.ExcludePatterns {
		match, err := filepath.Match(pattern, path)
		if err != nil {
			log.Println("excludeFile", err)
			continue
		}
		if match {
			return true
		}

	}
	return false
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
				log.Println("START:", f.LocalPath, *f.Params.Key)

				err := localToS3(s3Svc, f)
				if err != nil {
					log.Print(err)
				}

				log.Printf("DONE (%s): %s %s", time.Since(start), f.LocalPath, *f.Params.Key)
			}
		}()
	}

	filterPath := func(path string, info os.FileInfo) bool {
		if s.CopySymlinks && info.Mode()&os.ModeSymlink != 0 {
			return s.excludeFile(path)
		}
		if info.Mode().IsRegular() {
			return s.excludeFile(path)
		}

		return true
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if filterPath(path, info) {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		key := prefix + relPath

		if info.Mode().IsRegular() && bucketIndex.ExistsSize(key, info.Size()) {
			debug("Exists Size:", key)
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				log.Println(err)
				return nil
			}

			if bucketIndex.ExistsETAG(key, bytes.NewBufferString(target)) {
				debug("Exists ETAG:", key)
				return nil
			}
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

func md5Sum(src io.Reader) string {
	h := md5.New()
	io.Copy(h, src)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type S3KeyMap map[string]*s3.Object

func (s S3KeyMap) ExistsSize(key string, size int64) bool {
	v, ok := s[key]
	if !ok {
		return false
	}

	return *v.Size == size
}

func (s S3KeyMap) ExistsETAG(key string, src io.Reader) bool {
	v, ok := s[key]
	if !ok {
		return false
	}

	log.Println(key, *v.ETag, md5Sum(src))

	return *v.ETag == md5Sum(src)
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
		in.Params.Body = bytes.NewReader([]byte(target))

	} else if in.Info.Mode().IsRegular() {
		file, err := os.Open(in.LocalPath)
		if err != nil {
			return err
		}
		defer file.Close()

		in.Params.Body = file
	}

	if stat_t, ok := in.Info.Sys().(*syscall.Stat_t); ok {
		metadata["mode"] = aws.String(fmt.Sprint(stat_t.Mode))
		metadata["uid"] = aws.String(fmt.Sprint(stat_t.Uid))
		metadata["gid"] = aws.String(fmt.Sprint(stat_t.Gid))
	}

	if len(metadata) > 0 {
		in.Params.Metadata = &metadata
	}

	_, err := s3Svc.PutObject(in.Params)
	return err
}
