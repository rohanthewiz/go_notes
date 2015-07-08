package main

import(
	"fmt"
	"time"
	"crypto/sha1"
	"strings"
	"log"
)

// --- Some Shortcuts for often used output functions ---

var fpf = fmt.Printf
var fpl = fmt.Println
var lpl = log.Println

func pd(params ...interface{}) {
	if opts_intf["debug"].(bool) {
		log.Println(params...)
	}
}

func pl(params ...interface{}) {
	if opts_intf["verbose"].(bool) {
		fmt.Println(params...)
	}
}

func pf(msg string, params ...interface{}) {
	if opts_intf["verbose"].(bool) {
		fmt.Printf(msg, params...)
	}
}

// --- Crypto ---

func generate_sha1() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("%$" + time.Now().String() + "e{")))
}

func hashPassword(pword string, salt string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte("[--]" + pword + "e{" + salt)))

}

func short_sha(sha string) string{
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}


func trim_whitespace(in_str string) string {
	return strings.Trim(in_str, " \n\r\t")
}
