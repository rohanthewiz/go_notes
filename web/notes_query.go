package web

import (
	"go_notes/note"
	"go_notes/note/web"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rohanthewiz/rlog"
	"github.com/rohanthewiz/serr"
)

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

func WebListNotes(c *fiber.Ctx, nf *note.NotesFilter) (err error) {
	notes := note.QueryNotes(nf)
	err = web.NotesList(c, notes)
	if err != nil {
		return serr.Wrap(err, "Error in notes list html gen")
	}
	return
}
