package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_ops"
	"go_notes/note/web"
	"go_notes/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/julienschmidt/httprouter"
)

// Handlers for httprouter
func Index(c *fiber.Ctx) {
	c.Redirect("/q/all/l/100", http.StatusTemporaryRedirect)
}

func WebListNotes(w http.ResponseWriter, r *http.Request, nf *note.NotesFilter) {
	notes := note.QueryNotes(nf)

	err := web.NotesList(w, r, notes)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func Query(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	limit, err := strconv.Atoi(p.ByName("limit"))
	if err != nil {
		limit = 50
	}

	nf := note.NotesFilter{
		QueryStr: p.ByName("query"),
		Limit:    limit,
	}
	WebListNotes(w, r, &nf)
}

func QueryLast(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WebListNotes(w, r, &note.NotesFilter{Last: true})
}

func QueryId(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	WebListNotes(w, r, &note.NotesFilter{Id: id})
}

func QueryIdAsJson(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		id = 0
	}
	notes := note.QueryNotes(&note.NotesFilter{Id: id})
	jNotes, err := json.Marshal(notes)
	if err != nil {
		log.Println("Error marshalling Note id:", strconv.FormatInt(id, 10))
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jNotes)
	if err != nil {
		log.Println("Error in notes JSON gen:", err)
	}
}

func QueryTag(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tags := p.ByName("tag")
	WebListNotes(w, r, &note.NotesFilter{Tags: strings.Split(tags, ",")})
}

func QueryTitle(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	WebListNotes(w, r, &note.NotesFilter{Title: p.ByName("title")})
}

func QueryTagAndWildCard(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tags := strings.Split(p.ByName("tag"), ",")
	WebListNotes(w, r, &note.NotesFilter{Tags: tags, QueryStr: p.ByName("query")})
}

func QueryTitleAndWildCard(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	WebListNotes(w, r,
		&note.NotesFilter{Title: p.ByName("title"), QueryStr: p.ByName("query")})
}

func WebNoteForm(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		var nte note.Note
		dbhandle.DB.Where("id = ?", id).First(&nte) // get the original for comparision
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

	nb := strings.TrimSpace(v.Get("note_body"))
	nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

	tl := strings.TrimSpace(v.Get("title"))
	if tl == "" {
		HandleRequestErr(errors.New("title should not be empty"), w)
		return
	}

	id := note_ops.CreateNote(tl, strings.TrimSpace(v.Get("descr")),
		nb, strings.TrimSpace(v.Get("tag")))
	http.Redirect(w, r, "/qi/"+strconv.FormatUint(id, 10), http.StatusFound)
}

func WebNoteDup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	origId, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		HandleRequestErr(err, w)
		return
	}

	nte := note.FindNoteById(origId)
	if nte.Id < 1 {
		http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
	}

	// TODO - Check that note with title below does not already exist
	// 		and gracefully handle error
	id := note_ops.CreateNote("Copy of - "+nte.Title, "",
		"", nte.Tag)
	if id > 0 {
		http.Redirect(w, r, "/edit/"+strconv.FormatUint(id, 10), http.StatusFound)
	} else {
		http.Redirect(w, r, "/q/all/l/100", http.StatusFound)
	}
}

func WebDeleteNote(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	returnPath := "/q/all/l/100"
	id, err := strconv.ParseInt(p.ByName("id"), 10, 64)
	if err != nil {
		fmt.Println("Error parsing id when deleting note.")
		http.Redirect(w, r, returnPath, http.StatusFound)
		return
	}

	note.DoDelete(note.FindNoteById(id))

	qs := r.URL.Query()
	if qs != nil {
		returnPath = qs.Get("return")
	}

	http.Redirect(w, r, returnPath, http.StatusFound)
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

		nb := strings.TrimSpace(v.Get("note_body"))
		nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

		nte = note.Note{Id: id, Title: strings.TrimSpace(v.Get("title")),
			Description: strings.TrimSpace(v.Get("descr")),
			Body:        nb, Tag: strings.TrimSpace(v.Get("tag")),
		}

		utils.Pf("Updating note: %s %s...\n", nte.Guid, nte.Title)
		note.AllFieldsUpdate(nte)
		http.Redirect(w, r, "/qi/"+strconv.FormatUint(nte.Id, 10), http.StatusFound)
	}
}

func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
}

func HandleRequestErr(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprint(w, err)
}
