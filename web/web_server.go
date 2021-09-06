package web

import (
	"crypto/tls"
	"fmt"
	"go_notes/config"
	"log"
	"net/http"

	"github.com/gofiber/fiber"
	"github.com/rohanthewiz/rlog"
	"golang.org/x/crypto/acme/autocert"
)

func Webserver(port string) {
	const startMsg = "Web server listening on %s... Ctrl-C to quit"
	app := fiber.New()
	DoRoutes(app)

	if config.Opts.IsRemoteSvr {
		go func() { _ = http.ListenAndServe(":80", http.HandlerFunc(httpToHttps)) }()

		m := &autocert.Manager{
			Cache:      autocert.DirCache("."),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("gonotes.net"), // TODO - add domain to config
		}
		tcfg := &tls.Config{GetCertificate: m.GetCertificate}

		port = "443"
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.Listen(port, tcfg))
	} else {
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.Listen(port))
	}
}

func httpToHttps(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.String()
	fmt.Println("redirecting to ", target)
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
