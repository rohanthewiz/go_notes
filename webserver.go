package main
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
)
const listen_port string = "8080"

func doWebServer() {
	router := httprouter.New()
	doRoutes(router)
	pf("Server listening on %s... Ctrl-C to quit", listen_port)
	log.Fatal(http.ListenAndServe(":" + listen_port, router))
}

// Handlers for httprouter
func Index(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Query(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	// messing with sha1 //println(generate_sha1())
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes(opts_str, opts_intf )
	RenderQuery(w, notes) //call Ego generated method
}
