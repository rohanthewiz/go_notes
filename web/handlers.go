package web

import (
	"errors"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_ops"
	"go_notes/note/web"
	"go_notes/utils"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/julienschmidt/httprouter"
	"github.com/rohanthewiz/rlog"
	"github.com/rohanthewiz/serr"
)

// Handlers for httprouter
func Index(c *fiber.Ctx) error {
	err := c.Redirect("/q/all/l/100", http.StatusTemporaryRedirect)
	if err != nil {
		rlog.LogErr(err, "Error redirecting to index route")
	}
	return err
}

func WebListNotes(c *fiber.Ctx, nf *note.NotesFilter) {
	notes := note.QueryNotes(nf)

	err := web.NotesList(c, notes)
	if err != nil {
		log.Println("Error in notes list html gen:", err)
	}
}

func Query(c *fiber.Ctx) {
	limit, err := strconv.Atoi(c.Params("limit"))
	if err != nil {
		limit = 100
	}

	nf := note.NotesFilter{
		QueryStr: c.Params("query"),
		Limit:    limit,
	}
	WebListNotes(c, &nf)
}

func QueryLast(c *fiber.Ctx) error {
	WebListNotes(c, &note.NotesFilter{Last: true})
	return nil
}

func QueryId(c *fiber.Ctx) {
	id, err := c.ParamsInt("id")
	if err != nil {
		id = 0
	}
	WebListNotes(c, &note.NotesFilter{Id: int64(id)})
}

func QueryIdAsJson(c *fiber.Ctx) {
	id, err := c.ParamsInt("id")
	if err != nil {
		id = 0
	}
	notes := note.QueryNotes(&note.NotesFilter{Id: int64(id)})
	err = c.JSON(notes)
	if err != nil {
		rlog.LogErr(serr.Wrap(err, "Error in notes JSON gen:"))
	}
	// jNotes, err := json.Marshal(notes)
	// if err != nil {
	// 	log.Println("Error marshalling Note id:", strconv.Itoa(id))
	// }
	//
	// w.Header().Set("Content-Type", "application/json")
	// _, err = w.Write(jNotes)
}

func QueryTag(c *fiber.Ctx, p httprouter.Params) {
	tags := p.ByName("tag")
	WebListNotes(c, &note.NotesFilter{Tags: strings.Split(tags, ",")})
}

func QueryTitle(c *fiber.Ctx, p httprouter.Params) {
	WebListNotes(c, &note.NotesFilter{Title: p.ByName("title")})
}

func QueryTagAndWildCard(c *fiber.Ctx, p httprouter.Params) {
	tags := strings.Split(p.ByName("tag"), ",")
	WebListNotes(c, &note.NotesFilter{Tags: tags, QueryStr: p.ByName("query")})
}

func QueryTitleAndWildCard(c *fiber.Ctx, p httprouter.Params) {
	WebListNotes(c,
		&note.NotesFilter{Title: p.ByName("title"), QueryStr: p.ByName("query")})
}

func WebNoteForm(c *fiber.Ctx, p httprouter.Params) {
	if id, err := strconv.ParseInt(p.ByName("id"), 10, 64); err == nil {
		var nte note.Note
		dbhandle.DB.Where("id = ?", id).First(&nte) // get the original for comparision
		fmt.Printf("note at WebNoteForm %#v\n", nte.Guid)
		if nte.Id > 0 {
			err = web.NoteForm(c, nte)
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		} else {
			err := web.NoteForm(c, note.Note{})
			if err != nil {
				log.Println("Error in Render NoteForm:", err)
			}
		}
	} else {
		err := web.NoteForm(c, note.Note{})
		if err != nil {
			log.Println("Error in Render NoteForm:", err)
		}
	}
}

func WebCreateNote(c *fiber.Ctx) {
	postData := c.Body()

	v, err := url.ParseQuery(string(postData))
	if err != nil {
		HandleRequestErr(err, c)
		return
	}

	nb := strings.TrimSpace(v.Get("note_body"))
	nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

	tl := strings.TrimSpace(v.Get("title"))
	if tl == "" {
		HandleRequestErr(errors.New("title should not be empty"), c)
		return
	}

	id := note_ops.CreateNote(tl, strings.TrimSpace(v.Get("descr")),
		nb, strings.TrimSpace(v.Get("tag")))
	_ = c.Redirect("/qi/"+strconv.FormatUint(id, 10), http.StatusFound)
}

func WebNoteDup(c *fiber.Ctx) {
	origId, err := c.ParamsInt("id")
	if err != nil {
		HandleRequestErr(err, c)
		return
	}

	nte := note.FindNoteById(int64(origId))
	if nte.Id < 1 {
		_ = c.Redirect("/q/all/l/100", http.StatusFound)
	}

	// TODO - Check that note with title below does not already exist
	// 		and gracefully handle error
	id := note_ops.CreateNote("Copy of - "+nte.Title, "",
		"", nte.Tag)
	if id > 0 {
		_ = c.Redirect("/edit/"+strconv.FormatUint(id, 10), http.StatusFound)
	} else {
		_ = c.Redirect("/q/all/l/100", http.StatusFound)
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

func WebUpdateNote(c *fiber.Ctx) {
	var nte note.Note
	if id, err := c.ParamsInt("id"); err == nil {
		postData := c.Body()

		v, err := url.ParseQuery(string(postData))
		if err != nil {
			HandleRequestErr(err, c)
			return
		}

		nb := strings.TrimSpace(v.Get("note_body"))
		nb = note.UpsertKeyNotes(nb) // prepend KeyNotes - hardwired ON for now

		nte = note.Note{Id: uint64(id), Title: strings.TrimSpace(v.Get("title")),
			Description: strings.TrimSpace(v.Get("descr")),
			Body:        nb, Tag: strings.TrimSpace(v.Get("tag")),
		}

		utils.Pf("Updating note: %s %s...\n", nte.Guid, nte.Title)
		note.AllFieldsUpdate(nte)
		err = c.Redirect("/qi/"+strconv.FormatUint(nte.Id, 10), http.StatusFound)
		if err != nil {
			rlog.LogErr(err, "Error on redirect to note after update")
			_ = c.Redirect("/")
		}
	} else {
		rlog.LogErr(err, "No id found in update URL path")
	}
}

func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
}

func HandleRequestErr(err error, c *fiber.Ctx) {
	rlog.LogErr(err)
	// fhr := c.Response()
	// if fhr != nil {
	// 	//
	// }
	_ = c.SendStatus(http.StatusBadRequest)
}
