package web

import (
	"errors"
	"fmt"
	"go_notes/note"
	"go_notes/note/note_ops"
	"go_notes/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rohanthewiz/rlog"
)

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
