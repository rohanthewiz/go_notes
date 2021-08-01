package utils

import (
	"crypto/sha1"
	"fmt"
	"go_notes/config"
	"log"
	"time"
)

func Pd(params ...interface{}) {
	if config.Opts.Debug {
		log.Println(params...)
	}
}

func Pl(params ...interface{}) {
	if config.Opts.Verbose {
		fmt.Println(params...)
	}
}

func Pf(msg string, params ...interface{}) {
	if config.Opts.Verbose {
		fmt.Printf(msg, params...)
	}
}

func GenerateSHA1() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("%$"+time.Now().String()+"e{")))
}

func ShortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
