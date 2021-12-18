package web

import (
	"go_notes/dbhandle"
	"go_notes/note"
	"go_notes/note/web"
	"net/http"
	"strconv"

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

func WebNoteForm(c *fiber.Ctx) (err error) {
	defer func() {
		if err != nil {
			rlog.LogErr(err, "Error in WebNoteForm")
		}
	}()

	if id, err := strconv.ParseInt(c.Params("id"), 10, 64); err == nil {
		var nte note.Note
		dbhandle.DB.Where("id = ?", id).First(&nte) // get the original for comparison
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

func HandleRequestErr(err error, c *fiber.Ctx) {
	rlog.LogErr(err)
	_ = c.SendStatus(http.StatusBadRequest)
}
