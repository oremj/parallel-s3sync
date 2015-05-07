package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
	"github.com/oremj/parallel-s3sync/s3sync"
)

func main() {

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	app := cli.NewApp()
	app.Name = "parallel-s3sync"
	app.Usage = "<source> <target>"
	app.Version = "1.0.1"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "workers",
			Value: 16,
			Usage: "Set amount of parallel uploads",
		},
		cli.BoolFlag{
			Name:  "copy-symlinks",
			Usage: "copy, but do not follow symlinks",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "verbose logging",
		},
		cli.IntFlag{
			Name:  "loglevel",
			Value: 0,
			Usage: "Sets aws-sdk-go log level",
		},
		cli.StringSliceFlag{
			Name:  "exclude",
			Usage: "Matches based on http://golang.org/pkg/path/filepath/#Match",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "exclude-dir",
			Usage: "if the directory matches, skip. No trailing /",
			Value: &cli.StringSlice{},
		},
	}

	app.Action = func(c *cli.Context) {
		if len(c.Args()) < 2 {
			fmt.Println("<source> and <target> required")
			os.Exit(1)
		}
		source, target := c.Args()[0], c.Args()[1]
		workers := c.Int("workers")
		copySymlinks := c.Bool("copy-symlinks")
		s3sync.Debug = c.Bool("debug")
		excludes := c.StringSlice("exclude")
		excludeDirs := c.StringSlice("exclude-dir")

		s3Url, err := url.Parse(c.Args()[1])
		if s3Url.Scheme != "s3" || s3Url.Host == "" || s3Url.Path == "" {
			log.Fatal("Not a valid s3_path. Example: s3://bucket/path")
		}

		sync := s3sync.New(&aws.Config{
			MaxRetries: 5,
			LogLevel:   uint(c.Int("loglevel")),
		})
		for _, d := range excludeDirs {
			sync.ExcludeDirectories[d] = true
		}
		sync.ExcludePatterns = excludes
		sync.CopySymlinks = copySymlinks
		err = sync.Sync(source, target, workers)
		if err != nil {
			log.Fatal(err)
		}
	}

	app.Run(os.Args)
}
