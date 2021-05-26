package main

import (
	note2 "go_notes/note"
	"go_notes/note/web"
	"net/http"

	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/julienschmidt/httprouter"
	//	"github.com/microcosm-cc/bluemonday"
	"log"
	"net/url"
	"path"
	"strconv"
)

// Good reading: http://www.alexedwards.net/blog/golang-response-snippets

func webserver(listen_port string) {
	router := httprouter.New()
	doRoutes(router)
	pf("Server listening on %s... Ctrl-C to quit", listen_port)
	log.Fatal(http.ListenAndServe(":"+listen_port, router))
}

// Handlers for httprouter
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
	//	fmt.Fprint(w, "Welcome!\n")
}

func Query(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["q"] = p.ByName("query") // Overwrite the query param
	limit, err := strconv.Atoi(p.ByName("limit"))
	if err == nil {
		optsIntf["l"] = limit
	}
	notes := queryNotes()
	err = web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error building query response:", err)
	}
}

func QueryLast(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	resetOptions()
	optsIntf["ql"] = true // qi is the highest priority
	notes := queryNotes()
	err := web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func QueryId(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	optsIntf["qi"] = id // qi is the highest priority
	notes := queryNotes()
	err = web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	} //call Ego generated method
}

func QueryIdAsJson(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	optsIntf["qi"] = id // qi is the highest priority
	jNotes, err := json.Marshal(queryNotes())
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jNotes)
	if err != nil {
		log.Println("Error in notes JSON gen:", err)
	}
}

func QueryTag(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qg"] = p.ByName("tag") // Overwrite the query param
	optsIntf["qi"] = nil            // turn off unused option
	optsStr["qt"] = ""              // turn off unused option
	optsStr["q"] = ""               // turn off unused option
	notes := queryNotes()
	err := web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func QueryTitle(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qt"] = p.ByName("title") // Overwrite the query param
	optsIntf["qi"] = nil              // turn off unused option
	optsStr["qg"] = ""                // turn off unused option
	notes := queryNotes()
	err := web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func QueryTagAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qg"] = p.ByName("tag")  // Overwrite the query param
	optsStr["q"] = p.ByName("query") // Overwrite the query param
	notes := queryNotes()
	err := web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func QueryTitleAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qt"] = p.ByName("title") // Overwrite the query param
	optsStr["q"] = p.ByName("query")  // Overwrite the query param
	notes := queryNotes()
	err := web.NotesListDetailed(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func WebNoteForm(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		var note note2.Note
		db.Where("id = ?", id).First(&note) // get the original for comparision
		if note.Id > 0 {
			err := RenderNoteForm(w, note)
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		} else {
			err := RenderNoteForm(w, note2.Note{})
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		}
	} else {
		err := RenderNoteForm(w, note2.Note{})
		if err != nil {
			log.Println("Error in Render NoteForm:", err)
		}
	}
}

func WebCreateNote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HandleRequestErr(err, w)
		return
	}
	v, err := url.ParseQuery(string(post_data))
	if err != nil {
		HandleRequestErr(err, w)
		return
	}

	id := CreateNote(trimWhitespace(v.Get("title")), trimWhitespace(v.Get("description")),
		trimWhitespace(v.Get("body")), trimWhitespace(v.Get("tag")))
	http.Redirect(w, r, "/qi/"+strconv.FormatUint(id, 10), http.StatusFound)
}

func WebDeleteNote(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		fmt.Println("Error deleting note.")
	} else {
		DoDelete(findNoteById(id))
	}
	http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
}

func WebUpdateNote(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var note note2.Note
	if id, err := strconv.ParseUint(p.ByName("id"), 10, 64); err == nil {
		post_data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			HandleRequestErr(err, w)
			return
		}
		v, err := url.ParseQuery(string(post_data))
		if err != nil {
			HandleRequestErr(err, w)
			return
		}

		note = note2.Note{Id: id, Title: trimWhitespace(v.Get("title")),
			Description: trimWhitespace(v.Get("description")),
			Body:        trimWhitespace(v.Get("body")), Tag: trimWhitespace(v.Get("tag")),
		}
		pf("Updating note with: %v ...\n", note)
		AllFieldsUpdate(note)
		http.Redirect(w, r, "/qi/"+strconv.FormatUint(note.Id, 10), http.StatusFound)
	}
}

func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
}

func resetOptions() {
	optsIntf["qi"] = nil   // turn off unused option
	optsIntf["ql"] = false // turn off unused option
	optsIntf["l"] = -1     // turn off unused option
	optsStr["qg"] = ""     // turn off higher priority option
	optsStr["qt"] = ""     // turn off unused option
	optsStr["qd"] = ""     // turn off unused option
	optsStr["qb"] = ""     // turn off unused option
	optsStr["q"] = ""      // turn off higher priority option
}

func HandleRequestErr(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprint(w, err)
}

//	Only applies to GET request? //pl("p.Title", p.ByName("title"))
