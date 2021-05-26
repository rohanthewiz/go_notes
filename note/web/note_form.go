package web

import (
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"log"
	"strconv"

	"github.com/rohanthewiz/element"
)

func NoteForm(w io.Writer, note note.Note) (err error) {
	var action, button string
	if note.Id > 0 {
		action = "/note/" + strconv.FormatUint(note.Id, 10)
		button = "Update"
	} else {
		action = "/create/"
		button = "Create"
	}
	e := element.New
	str := e("html").R(
		e("head").R(
			e("title").R("Note Form"),
			e("style").R(`
	body { background-color: tan }
	.container { padding: 1em; border: 1px solid gray; border-radius: 0.5em }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #531C1C}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #8A2E2E; padding: 0.3em 0.3em}
    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em 4em }
    .h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }
    .title { font-weight: bold; color:darkgreen; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#401020; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }
    .note-body { padding-left:1.5em; margin-top: 0.1em}
	input { background-color: #EEE6D0 }
	textarea { background-color: #ECE6D0 }`),
		),
		e("body").R(
			e("h1").R("Note"),
			e("div", "class", "container").R(
				e("form", "action", action, "method", "post").R(
					e("table").R(
						e("tr").R(
							e("td").R(
								e("label", "for", "title").R("Title"),
								e("input", "name", "title", "type", "text", "size", "54", "value", html.EscapeString(note.Title)).R(),
							),
							e("td").R(
								e("label", "for", "tag").R("&nbsp;&nbsp;Tags"),
								e("input", "name", "tag", "type", "text", "size", "24", "value", html.EscapeString(note.Tag)).R(),
							),
						),
					),
					e("p").R(
						e("label", "for", "descr").R("Description"),
						e("input", "name", "descr", "size", "83", "value", html.EscapeString(note.Description)).R(),
					),
					e("p").R(
						e("label", "for", "body").R("Body"),
						e("textarea", "name", "body", "row", "14", "cols", "76").R(note.Body),
					),
					e("p").R(
						e("input", "type", "submit", "value", button).R(),
					),
				),
			),
		),
	)
	fmt.Println(str)

	_, err = fmt.Fprint(w, str)
	if err != nil {
		log.Println("Error on NoteForm render:", err)
	}

	return
}
