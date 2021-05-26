package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"strings"
	"time"
)

// --- Some Shortcuts for often used output functions ---

func pd(params ...interface{}) {
	if optsIntf["debug"].(bool) {
		log.Println(params...)
	}
}

func pl(params ...interface{}) {
	if optsIntf["verbose"].(bool) {
		fmt.Println(params...)
	}
}

func pf(msg string, params ...interface{}) {
	if optsIntf["verbose"].(bool) {
		fmt.Printf(msg, params...)
	}
}

// --- Crypto ---

func generateSHA1() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("%$"+time.Now().String()+"e{")))
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func trimWhitespace(in_str string) string {
	return strings.Trim(in_str, " \n\r\t")
}
