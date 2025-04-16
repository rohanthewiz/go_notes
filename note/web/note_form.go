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

	b, e, t := element.Vars()

	e("html").R(
		e("head").R(
			e("title").R(t("GoNotes Form")),
			e("style").R(t(string(noteFormStyles))),
			// We *must* load msgpack before monaco as the js loading is not happening after
			e("script", "src", "https://rawgithub.com/kawanet/msgpack-lite/master/dist/msgpack.min.js").R(),

			e("link", "rel", "stylesheet", "href", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs/editor/editor.main.css").R(),
			e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs/loader.js").R(),
		),

		e("body").R(
			e("span", "class", "h1").R(
				e("a", "href", "/").R(t("GoNotes  ")),
			),
			e("span", "class", "h1").R(t(pageHeadingPrefix, "Note")),
			e("div", "class", "container").R(
				e("form", "id", "note_form", "action", action, "method", "post").R(
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
					e("p", "style", "position: relative").R(
						e("label", "for", "note_body").R(t("Body (F1 for Cmd Palette)"), e("br")),
						e("div", "id", "editor").R(t("")),
						e("input", "type", "hidden", "id", "note_body", "name", "note_body").R(),
					),

					e("div", "class", "action-btns").R(
						e("p").R(
							e("input", "type", "submit", "class", "action-btn", "value", "Cancel", "formaction", "/"),
							b.Wrap(func() {
								if note.Id > 0 {
									e("input", "type", "submit", "class", "action-btn dup", "value", "Dup", "formaction", "/dup/"+strNoteId)
								}
							}),
							e("input", "type", "submit", "id", "create_update_btn", "class", "action-btn",
								"onsubmit", ";", "value", formAction), // the event handler here is being overridden by the note_form event handler in note_form.js
						),
					),
				),
			),

			e("script", "type", "text/javascript").R(
				t("document.addEventListener('DOMContentLoaded', function() {"),
				t(`var bin = atob('`, b64code, `');`),
				t(`window.codeObj = msgpack.decode(Uint8Array.from(bin, c => c.charCodeAt(0)));
						console.log(codeObj);`),
				t(string(noteFormJS)),
				t("});"), // close DOM Event listener
			),
		),
	)

	_, err = fmt.Fprint(w, b.String())
	if err != nil {
		log.Println("Error on NoteForm render:", err)
	}

	return
}
