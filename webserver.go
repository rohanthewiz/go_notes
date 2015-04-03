package main
import (
	"net/http"
	"github.com/julienschmidt/httprouter"
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
	resetOptions()
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	limit, err := strconv.Atoi(p.ByName("limit"))
	if err == nil {
		opts_intf["ql"] = limit
	}
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func QueryId(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil { id = 0 }
	opts_intf["qi"] = id  // qi is the highest priority
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func QueryTag(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	opts_str["qg"] = p.ByName("tag")  // Overwrite the query param
	opts_intf["qi"] = nil // turn off unused option
	opts_str["qt"] = "" // turn off unused option
	opts_str["q"] = "" // turn off unused option
	notes := queryNotes()
	RenderQuery(w, notes)
}

func QueryTitle(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	opts_str["qt"] = p.ByName("title")  // Overwrite the query param
	opts_intf["qi"] = nil // turn off unused option
	opts_str["qg"] = "" // turn off unused option
	notes := queryNotes()
	RenderQuery(w, notes)
}

func QueryTagAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	opts_str["qg"] = p.ByName("tag")  // Overwrite the query param
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func QueryTitleAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	opts_str["qt"] = p.ByName("title")  // Overwrite the query param
	opts_str["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes()
	RenderQuery(w, notes) //call Ego generated method
}

func WebNoteForm(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		var note Note
		db.Where("id = ?", id).First(&note) // get the original for comparision
		if note.Id > 0 {
			RenderNoteForm(w, note)
		} else {
			RenderNoteForm(w, Note{})
		}
	} else {
		RenderNoteForm(w, Note{})
	}
}

func WebCreateNote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil { HandleRequestErr(err, w); return }
	v, err := url.ParseQuery(string(post_data))
	if err != nil { HandleRequestErr(err, w); return }

	id := createNote(trim_whitespace(v.Get("title")), trim_whitespace(v.Get("description")),
		trim_whitespace(v.Get("body")), trim_whitespace(v.Get("tag")))
	http.Redirect(w, r, "/qi/" + strconv.FormatInt(id, 10), http.StatusFound)
}

func WebDeleteNote(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		println("Error deleting note.")
	} else {
		doDelete(find_note_by_id(id))
	}
	http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
}

func WebUpdateNote(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var note Note
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		post_data, err := ioutil.ReadAll(r.Body)
		if err != nil { HandleRequestErr(err, w); return }
		v, err := url.ParseQuery(string(post_data))
		if err != nil { HandleRequestErr(err, w); return }

		note = Note{ Id: id, Title: trim_whitespace(v.Get("title")),
			Description: trim_whitespace(v.Get("description")),
			Body: trim_whitespace(v.Get("body")),	Tag: trim_whitespace(v.Get("tag")),
		}
		pf("Updating note with: %v ...\n", note)
		allFieldsUpdate(note)
		http.Redirect(w, r, "/qi/"+strconv.FormatInt(note.Id, 10), http.StatusFound)
	}
}

func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
}

func resetOptions() {
	opts_intf["qi"] = nil // turn off unused option
	opts_intf["ql"] = -1 // turn off unused option
	opts_str["qg"] = "" // turn off higher priority option
	opts_str["qt"] = "" // turn off unused option
	opts_str["qd"] = "" // turn off unused option
	opts_str["qb"] = "" // turn off unused option
	opts_str["q"] = "" // turn off higher priority option
}

func HandleRequestErr(err error, w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
}

//	Only applies to GET request? //println("p.Title", p.ByName("title"))
