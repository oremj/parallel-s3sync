package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
	"github.com/oremj/parallel-s3sync/sync"
)

func main() {

	app := cli.NewApp()
	app.Name = "parallel-s3sync"
	app.Usage = "<local_path> <s3_path>"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "workers",
			Value: 16,
			Usage: "Set amount of parallel uploads",
		},
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) < 2 {
			fmt.Println("<local_path> and <s3_path> required")
			os.Exit(1)
		}
		workers := c.Int("workers")

		s3Url, err := url.Parse(c.Args()[1])
		if s3Url.Scheme != "s3" || s3Url.Host == "" || s3Url.Path == "" {
			log.Fatal("Not a valid s3_path. Example: s3://bucket/path")
		}

		if err != nil {
			log.Fatal(err)
		}
		localPath := c.Args()[0]

		syncer := &sync.Sync{
			AWSConfig: &aws.Config{},
			Bucket:    s3Url.Host,
		}
		syncer.Sync(localPath, s3Url.Path, workers)
	}

	app.Run(os.Args)
}
