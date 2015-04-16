package sync

import (
	"fmt"
	"log"
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
}

type Sync struct {
	AWSConfig   *aws.Config
	Bucket      string
	filesToSync []*syncFile
	position    int
	lk          sync.Mutex
}

type S3KeyMap map[string]*s3.Object

func (s S3KeyMap) Exists(key string, size int64) bool {
	v, ok := s[key]
	if !ok {
		return false
	}

	return *v.Size == size
}

func (s *Sync) nextSyncFile() (*syncFile, bool) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if s.position >= len(s.filesToSync) {
		return nil, false
	}

	f := s.filesToSync[s.position]
	s.position++

	return f, true

}

func (s *Sync) worker(wg *sync.WaitGroup) {
	s3Svc := s3.New(s.AWSConfig)
	for {
		f, more := s.nextSyncFile()
		if !more {
			wg.Done()
			return
		}

		err := s.PutFile(s3Svc, f.LocalPath, f.Key)
		if err != nil {
			log.Print(err)
		}
	}

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

func (s *Sync) PutFile(s3Svc *s3.S3, localPath, key string) error {
	fmt.Println("Putting:", localPath, key)
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

func (s *Sync) Sync(localPath, remotePath string, workers int) error {
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

		s.filesToSync = append(s.filesToSync, &syncFile{Key: key, LocalPath: path})
		return err
	})

	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go s.worker(wg)
	}

	wg.Wait()

	return nil
}
