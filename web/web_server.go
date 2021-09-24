package web

import (
	"crypto/tls"
	"fmt"
	"go_notes/config"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rohanthewiz/rlog"
	"golang.org/x/crypto/acme/autocert"
)

func Webserver(port string) {
	const startMsg = "Web server listening on %s... Ctrl-C to quit"
	app := fiber.New()
	DoRoutes(app) // TODO add static file server to `dist/`

	if config.Opts.IsRemoteSvr { // Create secure server
		go func() { _ = http.ListenAndServe(":80", http.HandlerFunc(httpToHttps)) }()

		m := &autocert.Manager{
			Cache:      autocert.DirCache("."),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("gonotes.net"), // TODO - add domain to config
		}
		tcfg := &tls.Config{GetCertificate: m.GetCertificate}

		app.Config()

		port = "443"
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		app.
		log.Fatal(app.ListenTLS(port, tcfg))

	} else { // local server - no auth
		rlog.Log(rlog.Info, fmt.Sprintf(startMsg, port))
		log.Fatal(app.Listen(port))
	}
}

func httpToHttps(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.String()
	fmt.Println("redirecting to ", target)
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
func whatever() { // TODO seeing if we can have static server for ACME and http->https redirector
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})


	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
	// log.Fatal(http.ListenAndServe(":8081", nil))
}