package main
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
//	"github.com/russross/blackfriday"
//	"github.com/microcosm-cc/bluemonday"
	"log"
	"fmt"
	"io/ioutil"
	"net/url"
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
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func WebNoteForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	fmt.Fprint(w, string("HI There!"))
	RenderShowNote(w, Note{Title: "Enter a title"})  //call Ego generated method
}

func WebCreateNote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	v, _ := url.ParseQuery(string(post_data))
//	Only applies to GET request //println("p.Title", p.ByName("title"))
//	println("p.Body", p.ByName("body"))

	fmt.Fprintf(w, "HI There!\nTitle: %s", v.Get("title"))
	println("Body:", v.Get("body"))
}

