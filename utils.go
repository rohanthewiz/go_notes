package main

import(
	"fmt"
	"time"
	"crypto/sha1"
)

func generate_sha1() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("%$" + time.Now().String() + "e{")))
}

func short_sha(sha string) string{
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
