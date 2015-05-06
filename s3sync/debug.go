package s3sync

import "log"

var Debug = false

func debug(v ...interface{}) {
	if Debug {
		log.Println(v...)
	}
}
