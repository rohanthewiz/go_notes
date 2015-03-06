package main
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
)

func doWebServer() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/q/:query", Query)
	println("Server listening on 8080... Ctrl-C to quit")
	log.Fatal(http.ListenAndServe(":8080", router))
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
