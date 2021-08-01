package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"go_notes/config"
	"log"
	"time"

	"golang.org/x/crypto/blake2b"
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

// Adapted from Beego
func RandomToken(ln int) (token string, err error) {
	b := make([]byte, ln)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		estr := "Could not read from the system CSPRNG - "
		if err != nil {
			estr += err.Error()
		}
		return token, errors.New(estr)
	}
	return hex.EncodeToString(b), nil
}

func RandomTokenBase64(hexLen int) (token string, err error) {
	b := make([]byte, hexLen)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		estr := "Could not read from the system CSPRNG - "
		if err != nil {
			estr += err.Error()
		}
		return token, errors.New(estr)
	}
	return base64.URLEncoding.EncodeToString(b), err
}

func Blake256(data string) string {
	h := blake2b.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:])
}

func Blake384(data string) string {
	h := blake2b.Sum384([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:])
}

func Blake512(data string) string {
	h := blake2b.Sum512([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:])
}
