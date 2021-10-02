package web

import (
	"errors"
	"fmt"
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/note_ops"
	"go_notes/note/web"
	"go_notes/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
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

func WebListNotes(c *fiber.Ctx, nf *note.NotesFilter) (err error) {
	notes := note.QueryNotes(nf)
	err = web.NotesList(c, notes)
	if err != nil {
		return serr.Wrap(err, "Error in notes list html gen")
	}
	return
}

func Query(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in web query")
		}
	}()

	limit, err := strconv.Atoi(c.Params("limit"))
	if err != nil {
		limit = 100
	}

	nf := note.NotesFilter{
		QueryStr: c.Params("query"),
		Limit:    limit,
	}
	err = WebListNotes(c, &nf)
	return
}

func QueryLast(c *fiber.Ctx) (err error) {
	err = WebListNotes(c, &note.NotesFilter{Last: true})
	if err != nil {
		rlog.LogErr(err, "Error in web queryLast")
	}
	return
}

func QueryId(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in web query")
		}
	}()
	id, err := c.ParamsInt("id")
	if err != nil {
		return err
	}
	err = WebListNotes(c, &note.NotesFilter{Id: int64(id)})
	return
}

func QueryIdAsJson(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in QueryIdAsJson")
		}
	}()

	id, err := c.ParamsInt("id")
	if err != nil {
		return serr.Wrap(err, "Unable to parse id")
	}

	notes := note.QueryNotes(&note.NotesFilter{Id: int64(id)})
	err = c.JSON(notes)
	if err != nil {
		return serr.Wrap(err, "Error in notes JSON gen:")
	}
	return
}

func QueryTag(c *fiber.Ctx) (err error) {
	tags := c.Params("tag")
	err = WebListNotes(c, &note.NotesFilter{Tags: strings.Split(tags, ",")})
	if err != nil {
		rlog.LogErr(err, "Error in web query by tag")
	}
	return
}

func QueryTitle(c *fiber.Ctx) (err error) {
	err = WebListNotes(c, &note.NotesFilter{Title: c.Params("title")})
	if err != nil {
		rlog.LogErr(err, "Error in web query by tag")
	}
	return
}

func QueryTagAndWildCard(c *fiber.Ctx) (err error) {
	tags := strings.Split(c.Params("tag"), ",")
	err = WebListNotes(c, &note.NotesFilter{Tags: tags, QueryStr: c.Params("query")})
	if err != nil {
		rlog.LogErr(err, "Error in query by tag and wildcard")
	}
	return
}

func QueryTitleAndWildCard(c *fiber.Ctx) (err error) {
	err = WebListNotes(c, &note.NotesFilter{Title: c.Params("title"), QueryStr: c.Params("query")})
	if err != nil {
		rlog.LogErr(err, "Error in query title and wildcard")
	}
	return
}

func WebNoteForm(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in WebNoteForm")
		}
	}()

	if id, err := strconv.ParseInt(c.Params("id"), 10, 64); err == nil {
		var nte note.Note
		dbhandle.DB.Where("id = ?", id).First(&nte) // get the original for comparision
		rlog.Log(rlog.Debug, "note at WebNoteForm: "+nte.Guid)

		var n note.Note
		if nte.Id > 0 {
			n = nte
		}

		err := web.NoteForm(c, n)
		if err != nil {
			return serr.Wrap(err, "Error in Render of NoteForm")
		}
	} else {
		err := web.NoteForm(c, note.Note{})
		if err != nil {
			return serr.Wrap(err)
		}
	}
	return
}

func WebCreateNote(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in WebCreateNote")
		}
	}()
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

	err = c.Redirect("/qi/"+strconv.FormatUint(id, 10), http.StatusFound)
	return
}

func WebNoteDup(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in WebNoteDup")
		}
	}()

	origId, err := c.ParamsInt("id")
	if err != nil {
		HandleRequestErr(err, c)
		return
	}

	nte := note.FindNoteById(int64(origId))
	if nte.Id < 1 {
		err = c.Redirect("/q/all/l/100", http.StatusFound)
	}

	// TODO - Check that note with title below does not already exist
	// 		and gracefully handle error
	id := note_ops.CreateNote("Copy of - "+nte.Title, "",
		"", nte.Tag)
	if id > 0 {
		err = c.Redirect("/edit/"+strconv.FormatUint(id, 10), http.StatusFound)
	} else {
		err = c.Redirect("/q/all/l/100", http.StatusFound)
	}
	return
}

func WebDeleteNote(c *fiber.Ctx) (err error) {
	var returnPath = "/q/all/l/100"

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		fmt.Println("Error parsing id when deleting note.")
		_ = c.Redirect(returnPath, http.StatusFound)
		return
	}

	note.DoDelete(note.FindNoteById(id))

	qs := c.Query("return")
	if qs != "" {
		returnPath = qs
	}

	return c.Redirect(returnPath, http.StatusFound)
}

func WebUpdateNote(c *fiber.Ctx) (err error) {
	var nte note.Note
	if id, err := c.ParamsInt("id"); err == nil {
		postData := c.Body()

		v, err := url.ParseQuery(string(postData))
		if err != nil {
			HandleRequestErr(err, c)
			return err
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
	return
}

// func ServeJS(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
// 	http.ServeFile(w, r, path.Join("js", p.ByName("file")))
// }

func HandleRequestErr(err error, c *fiber.Ctx) {
	rlog.LogErr(err)
	// fhr := c.Response()
	// if fhr != nil {
	// 	//
	// }
	_ = c.SendStatus(http.StatusBadRequest)
}
