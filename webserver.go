package main
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
//	"github.com/russross/blackfriday"
//	"github.com/microcosm-cc/bluemonday"
	"log"
	"fmt"
	"io/ioutil"
	"strconv"
	"net/url"
	"path"
)
const listen_port string = "8080"

// Good reading: http://www.alexedwards.net/blog/golang-response-snippets

func webserver() {
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
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func QueryById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)  // Overwrite the query param
	if err != nil { id = 0 }
	opts_intf["qi"] = id
	notes := queryNotes()
	RenderQuery(w, notes)
}

func WebNoteForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	RenderNoteForm(w, Note{})  //call Ego generated method
}

func WebCreateNote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil { HandleRequestErr(err, w); return }
	v, err := url.ParseQuery(string(post_data))
	if err != nil { HandleRequestErr(err, w); return }

	id := createNote(v.Get("title"), v.Get("description"), v.Get("body"), v.Get("tag"))
	println("Title:", v.Get("title"))
//	println("Title via FormValue:", r.FormValue("title"))
	http.Redirect(w, r, "/qi/" + strconv.FormatInt(id, 10), http.StatusFound)
}

func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
}

func HandleRequestErr(err error, w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
}

//	Only applies to GET request //println("p.Title", p.ByName("title"))
