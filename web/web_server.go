package web

import (
	"fmt"
	"go_notes/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rohanthewiz/rlog"
)

const tlsPort = ":443"

var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

func Webserver(port string) {
	const startMsg = "Web server listening on %s... Ctrl-C to quit"

	app := fiber.New()
	DoRoutes(app)
	// TODO add static file server to `dist/`

	if config.Opts.IsRemoteSvr { // Create secure server
		if cp := os.Getenv("GN_CERT_PATH"); cp != "" {
			certPath = cp
		}
		if kp := os.Getenv("GN_KEY_PATH"); kp != "" {
			keyPath = kp
		}
		// HTTP handling
		go func() { _ = http.ListenAndServe(":80", http.HandlerFunc(handleHttp)) }()

		// HTTPS
		port = ":443"
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.ListenTLS(tlsPort, certPath, keyPath))

	} else { // LOCAL server - no auth
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.Listen(port))
	}
}

// handleHttp handles the ACME challenge and forwards all other http requests to https
func handleHttp(w http.ResponseWriter, r *http.Request) {
	// Serve Challenge Certs
	if strings.Contains(r.URL.Path, "well-known/acme-challenge") {
		http.ServeFile(w, r, filepath.Join("./certs", r.URL.Path))
		return
	}
	// Forwarder
	target := "https://" + r.Host + r.URL.String()
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
