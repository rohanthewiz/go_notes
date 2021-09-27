package web

import (
	"fmt"
	"go_notes/config"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rohanthewiz/rlog"
)

const tlsPort = ":443"
const certPath = "./certs/.well_known/acme-challenge/cert.pem" // TODO establish paths
const keyPath = "./certs/.well_known/acme-challenge/key.pem"

func Webserver(port string) {
	const startMsg = "Web server listening on %s... Ctrl-C to quit"

	app := fiber.New()
	DoRoutes(app)
	// TODO add static file server to `dist/`

	if config.Opts.IsRemoteSvr { // Create secure server
		go func() { _ = http.ListenAndServe(":80", http.HandlerFunc(handleHttp)) }()

		port = ":443"
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.ListenTLS(tlsPort, certPath, keyPath))

	} else { // local server - no auth
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.Listen(port))
	}
}

// handleHttp handles the ACME challenge and forwards all other http requests to https
func handleHttp(w http.ResponseWriter, r *http.Request) {
	// Static server
	if strings.Contains(r.URL.Path, "well-known/acme-challenge") {
		http.ServeFile(w, r, filepath.Join("./certs", r.URL.Path))
		return
	}
	// Forwarder
	target := "https://" + r.Host + r.URL.String()
	fmt.Println("redirecting to ", target)
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
