# parallel-s3sync

```
NAME:
   parallel-s3sync - <source> <target>

USAGE:
   parallel-s3sync [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --workers "16"                       Set amount of parallel uploads
   --copy-symlinks                      copy, but do not follow symlinks
   --debug                          verbose logging
   --loglevel "0"                       Sets aws-sdk-go log level
   --exclude [--exclude option --exclude option]        Matches based on http://golang.org/pkg/path/filepath/#Match
   --exclude-dir [--exclude-dir option --exclude-dir option]    if the directory matches, skip. No trailing /
   --help, -h                           show help
   --version, -v                        print the version
```
