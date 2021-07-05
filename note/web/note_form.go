package web

import (
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/rohanthewiz/element"
)

func NoteForm(w io.Writer, note note.Note) (err error) {
	var action, formAction, pageHeadingPrefix string
	var strNoteId string

	if note.Id > 0 {
		strNoteId = strconv.FormatUint(note.Id, 10)
		action = "/note/" + strNoteId
		formAction = "Update"
		pageHeadingPrefix = "Edit "
	} else {
		action = "/create/"
		formAction = "Create"
		pageHeadingPrefix = "New "
	}

	s := &strings.Builder{}
	e := func(el string, p ...string) element.Element {
		return element.New(s, el, p...)
	}
	t := func(p ...string) int {
		return element.Text(s, p...)
	}

	e("html").R(
		e("head").R(
			e("title").R(t("GoNotes Form")),
			e("style").R(t(`
	body { background-color: tan }
	.container { padding: 1em; border: 1px solid gray; border-radius: 0.5em }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #531C1C}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #8A2E2E; padding: 0.3em 0.3em}
    td label {margin-right: 0.4em; font-size: 0.9em; color: #303030 }
    p label {margin-right: 0.4em; vertical-align: top; font-size: 0.9em; color: #303030}
    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em 4em }
    .h1 { font-size: 1.2em; margin-right: 0.2em; margin-bottom: 0.1em; padding: 0.1em }
	.h1 a {text-decoration:none}
	.h1 a:visited, .h1 a:link {color:black}
    .h3 { font-size: 1em; font-weight:bold; margin-bottom: 0.1em; padding: 0.1em }
    .title { font-size:1.1em; font-weight: bold; color:darkgreen; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#401020; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }
	.descr { width:99% }
    .note-body { padding-left:1.5em; margin-top: 0.1em; width:99%}
	button {cursor: pointer; margin: 0.5em 0.1em; vertical-align: baseline;}
	td input { margin-right: 0.8em; width:96% }
	.action-btns { text-align: right }
	input.action-btn { width: 10em; padding-left: 0.2em; padding-right: 0.2em; margin-right: 2em }
	textarea { background-color: #ECE6D0 }`)),
		),
		e("body").R(
			e("span", "class", "h1").R(
				e("a", "href", "/").R(t("GoNotes  ")),
			),
			e("span", "class", "h1").R(t(pageHeadingPrefix, "Note")),
			e("div", "class", "container").R(
				e("form", "action", action, "method", "post").R(
					// careful not to change any name attributes below, or form may break
					e("table").R(
						e("tr").R(
							e("td").R(
								e("label", "for", "title").R(t("Title")),
								e("input", "name", "title", "type", "text", "size", "54", "value", html.EscapeString(note.Title)),
							),
							e("td").R(
								e("label", "for", "tag").R(t("&nbsp;&nbsp;Tags")),
								e("input", "name", "tag", "type", "text", "size", "24", "value", html.EscapeString(note.Tag)),
							),
						),
					),
					e("p").R(
						e("label", "for", "descr").R(t("Description")),
						e("input", "class", "descr", "name", "descr", "size", "83", "value", html.EscapeString(note.Description)),
					),
					e("p").R(
						e("label", "for", "note_body").R(t("Body"), e("br")),
						e("textarea", "class", "note-body", "name", "note_body", "rows", "20").R(t(note.Body)),
					),
					e("div", "class", "action-btns").R(
						e("p").R(
							func() (r int) {
								if note.Id > 0 {
									e("input", "type", "submit", "class", "action-btn", "value", "Dup", "formaction", "/dup/"+strNoteId)
									// e("button", "onclick", "javascript:window.location='/duplicate/"+strNoteId+"'").R(t("Duplicate"))
								}
								return
							}(),
							e("input", "type", "submit", "class", "action-btn", "value", "Cancel", "formaction", "/"),
							e("input", "type", "submit", "class", "action-btn", "value", formAction),
						),
					),
				),
			),
		),
	)
	// fmt.Println(str)

	_, err = fmt.Fprint(w, s.String())
	if err != nil {
		log.Println("Error on NoteForm render:", err)
	}

	return
}
