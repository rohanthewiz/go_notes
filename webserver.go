package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_notes/note"
	"go_notes/note/web"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func webserver(listen_port string) {
	router := httprouter.New()
	doRoutes(router)
	pf("Server listening on %s... Ctrl-C to quit", listen_port)
	log.Fatal(http.ListenAndServe(":"+listen_port, router))
}

// Handlers for httprouter
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
}

func webListNotes(w http.ResponseWriter) {
	notes := queryNotes()

	err := web.NotesList(w, notes, optsStr)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func Query(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["q"] = p.ByName("query") // Overwrite the query param
	limit, err := strconv.Atoi(p.ByName("limit"))
	if err == nil {
		optsIntf["l"] = limit
	}
	webListNotes(w)
}

func QueryLast(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	resetOptions()
	optsIntf["ql"] = true
	webListNotes(w)
}

func QueryId(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	optsIntf["qi"] = id // qi is the highest priority
	webListNotes(w)
}

func QueryIdAsJson(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	optsIntf["qi"] = id // qi is the highest priority
	jNotes, err := json.Marshal(queryNotes())
	if err != nil {
		log.Println("Error marshalling Note id:", strconv.FormatInt(id, 10))
	}
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
	webListNotes(w)
}

func QueryTitle(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qt"] = p.ByName("title") // Overwrite the query param
	optsIntf["qi"] = nil              // turn off unused option
	optsStr["qg"] = ""                // turn off unused option
	webListNotes(w)
}

func QueryTagAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qg"] = p.ByName("tag")  // Overwrite the query param
	optsStr["q"] = p.ByName("query") // Overwrite the query param
	webListNotes(w)
}

func QueryTitleAndWildCard(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	resetOptions()
	optsStr["qt"] = p.ByName("title") // Overwrite the query param
	optsStr["q"] = p.ByName("query")  // Overwrite the query param
	webListNotes(w)
}

func WebNoteForm(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		var nte note.Note
		db.Where("id = ?", id).First(&nte) // get the original for comparision
		fmt.Printf("note at WebNoteForm %#v\n", nte.Guid)
		if nte.Id > 0 {
			err = web.NoteForm(w, nte)
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		} else {
			err := web.NoteForm(w, note.Note{})
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		}
	} else {
		err := web.NoteForm(w, note.Note{})
		if err != nil {
			log.Println("Error in Render NoteForm:", err)
		}
	}
}

func WebCreateNote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	postData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HandleRequestErr(err, w)
		return
	}

	v, err := url.ParseQuery(string(postData))
	if err != nil {
		HandleRequestErr(err, w)
		return
	}

	nb := trimWhitespace(v.Get("note_body"))
	nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

	tl := trimWhitespace(v.Get("title"))
	if tl == "" {
		HandleRequestErr(errors.New("title should not be empty"), w)
		return
	}

	id := CreateNote(tl, trimWhitespace(v.Get("descr")),
		nb, trimWhitespace(v.Get("tag")))
	http.Redirect(w, r, "/qi/"+strconv.FormatUint(id, 10), http.StatusFound)
}

func WebNoteDup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	origId, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		HandleRequestErr(err, w)
		return
	}

	nte := findNoteById(origId)
	if nte.Id < 1 {
		http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
	}

	// TODO - Check that note with title below does not already exist
	// 		and gracefully handle error
	id := CreateNote("Copy of - "+nte.Title, "",
		"", nte.Tag)
	if id > 0 {
		http.Redirect(w, r, "/edit/"+strconv.FormatUint(id, 10), http.StatusFound)
	} else {
		http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
	}
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
	var nte note.Note
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

		nb := trimWhitespace(v.Get("note_body"))
		nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

		nte = note.Note{Id: id, Title: trimWhitespace(v.Get("title")),
			Description: trimWhitespace(v.Get("descr")),
			Body:        nb, Tag: trimWhitespace(v.Get("tag")),
		}

		pf("Updating note with: %v ...\n", nte)
		AllFieldsUpdate(nte)
		http.Redirect(w, r, "/qi/"+strconv.FormatUint(nte.Id, 10), http.StatusFound)
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
