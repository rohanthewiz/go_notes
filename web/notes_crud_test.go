package web

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func TestWebReadNote(t *testing.T) {
	app := fiber.New()
	LoadRoutes(app)

	req := httptest.NewRequest("GET", "/q/all", nil)
	req.Header.Set("content-type", "text/html") // "application/json"

	resp, err := app.Test(req)
	if err != nil {
		t.Error("Error running query by index test - ", err)
	}

	if resp.StatusCode < 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error("error reading response body: ", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if len(body) < 50 { // TODO come up with correct logic here
			t.Error("response is shorter than expected")
		}
		fmt.Println(string(body))
		// fmt.Println(string(body)[:160])
	}
}

func TestWebCreateNote(t *testing.T) {
	app := fiber.New()
	LoadRoutes(app)

	data := url.Values{
		"title":     {"Title 2 of the test note"}, // TODO use rand for title
		"descr":     {"Description"},
		"note_body": {"Body of test the note"},
		"tag":       {"test, testing"},
	}

	req := httptest.NewRequest("POST", "/create", strings.NewReader(data.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Error("Error running create note test - ", err)
	}

	if resp.StatusCode < 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error("error reading response body: ", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		// if len(body) < 50 {
		// 	t.Error("response is shorter than expected")
		// }
		fmt.Println(string(body))
	}
}
