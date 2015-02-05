package main

import(
	"fmt"
	"time"
	"crypto/sha1"
)

func generate_sha1() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("%$" + time.Now().String() + "e{")))
}

