package web

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
)

func TestWebCreateNote(t *testing.T) {
	app := fiber.New()
	LoadRoutes(app)

	req := httptest.NewRequest("GET", "/qi/10", nil)
	req.Header.Set("content-type", "text/html") //"application/json"

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

		if len(body) < 50 {
			t.Error("response is shorter than expected")
		}
		fmt.Println(string(body)[:160])
	}
}
