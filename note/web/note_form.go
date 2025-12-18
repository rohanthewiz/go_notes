package web

import (
	"embed"
	"encoding/base64"
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"log"
	"strconv"

	"github.com/rohanthewiz/element"
	"github.com/rohanthewiz/serr"
	"github.com/vmihailenco/msgpack/v5"
)

//go:embed embeds
var embedFS embed.FS

func NoteForm(w io.Writer, note note.Note) (err error) {
	var action, formAction, pageHeadingPrefix string
	var strNoteId string

	noteFormStyles, err := embedFS.ReadFile("embeds/note_form.css")
	if err != nil {
		return serr.Wrap(err, "failed to load embedded note_form.css")
	}

	noteFormJS, err := embedFS.ReadFile("embeds/note_form.js")
	if err != nil {
		return serr.Wrap(err, "failed to load embedded note_form.js")
	}

	msgpackJS, err := embedFS.ReadFile("embeds/msgpack.min.js")
	if err != nil {
		return serr.Wrap(err, "failed to load embedded msgpack.min.js")
	}

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

	// CodeXfr is used to "escape" code to be sent to JS
	type CodeXfr struct {
		Code string `json:"code"`
	}

	mpkCode, err := msgpack.Marshal(CodeXfr{note.Body})
	if err != nil {
		return serr.Wrap(err, "Unable to marshal note body for frontend")
	}

	b64code := base64.StdEncoding.EncodeToString(mpkCode)
	// fmt.Println("**-> b64code", b64code)

	b := element.B()

	b.Html().R(
		b.Head().R(
			b.Title().T("GoNotes Form"),
			b.Style().T(string(noteFormStyles)),
			// We *must* load msgpack before monaco as the js loading is not happening after
			b.Script("type", "text/javascript").T(string(msgpackJS)),

			b.Link("rel", "stylesheet", "href", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs/editor/editor.main.css").R(),
			b.Script("type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs/loader.js").R(),
		),

		b.Body().R(
			b.SpanClass("h1").R(
				b.A("href", "/").T("GoNotes  "),
			),
			b.SpanClass("h1").T(pageHeadingPrefix, "Note"),
			b.DivClass("container").R(
				b.Form("id", "note_form", "action", action, "method", "post").R(
					// careful not to change any name attributes below, or form may break
					b.Table().R(
						b.Tr().R(
							b.Td().R(
								b.Label("for", "title").T("Title"),
								b.Input("name", "title", "type", "text", "size", "54", "value", html.EscapeString(note.Title)),
							),
							b.Td().R(
								b.Label("for", "tag").T("&nbsp;&nbsp;Tags"),
								b.Input("name", "tag", "type", "text", "size", "24", "value", html.EscapeString(note.Tag)),
							),
						),
					),

					b.P().R(
						b.Label("for", "descr").T("Description"),
						b.InputClass("descr", "name", "descr", "size", "83", "value", html.EscapeString(note.Description)),
					),
					b.P("style", "position: relative").R(
						b.Label("for", "note_body").R(
							b.T("Body (F1 for Cmd Palette)"),
							b.Br(),
						),
						b.Div("id", "editor").R(),
						b.Input("type", "hidden", "id", "note_body", "name", "note_body").R(),
					),

					b.DivClass("action-btns").R(
						b.P().R(
							b.InputClass("action-btn", "type", "submit", "value", "Cancel", "formaction", "/"),
							b.Wrap(func() {
								if note.Id > 0 {
									b.InputClass("action-btn dup", "type", "submit", "value", "Dup", "formaction", "/dup/"+strNoteId)
								}
							}),
							b.InputClass("action-btn", "type", "submit", "id", "create_update_btn",
								"onsubmit", ";", "value", formAction), // the event handler here is being overridden by the note_form event handler in note_form.js
						),
					),
				),
			),

			b.Script("type", "text/javascript").R(
				b.T("document.addEventListener('DOMContentLoaded', function() {"),
				b.T(`var bin = atob('`, b64code, `');`),
				b.T(`window.codeObj = msgpack.decode(Uint8Array.from(bin, c => c.charCodeAt(0)));
						console.log(codeObj);`),
				b.T(string(noteFormJS)),
				b.T("});"), // close DOM Event listener
			),
		),
	)

	_, err = fmt.Fprint(w, b.String())
	if err != nil {
		log.Println("Error on NoteForm render:", err)
	}

	return
}
