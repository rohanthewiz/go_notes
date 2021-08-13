package session

import (
	"fmt"
	"go_notes/utils"
	"net/http"
	"net/url"
	"time"
)

const sessionCookieName = "go_notes_session"
const cookieMaxAge = 60 * 3600

// Get our cookie from the request or generate a new one
func SessionStart(w http.ResponseWriter, r *http.Request) (sessId string, err error) {
	sessId, err = GetSessionIdFromRequestCookie(r)
	if err == nil && sessId != "" {
		return
	}

	return RegenerateSessionId(w, r)
}

func RegenerateSessionId(w http.ResponseWriter, r *http.Request) (sessId string, err error) {
	sessId, err = utils.RandomTokenBase64(64)
	if err != nil {
		fmt.Println("Error generating session", err.Error())
		return
	}

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    url.QueryEscape(sessId),
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		MaxAge:   cookieMaxAge,
		Expires:  time.Now().Add(time.Duration(cookieMaxAge) * time.Second),
	}

	http.SetCookie(w, cookie) // tell the client to set it on rcx of response
	r.AddCookie(cookie)       // set it as if we got it from the client
	return
}

func GetSessionIdFromRequestCookie(r *http.Request) (sid string, err error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return url.QueryUnescape(cookie.Value)
}
