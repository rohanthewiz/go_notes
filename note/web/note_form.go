package web

import (
	_ "embed"
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/rohanthewiz/element"
)

var (
	//go:embed embed/note_form.css
	noteFormStyles []byte
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
			e("style").R(t(string(noteFormStyles))),
		),
		e("link", "rel", "stylesheet", "href", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.0/min/vs/editor/editor.main.css"),
		e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.0/min/vs/loader.js").R(),

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
						// e("button", "id", "showCmdPal").R(t("Show Cmd Palette")),
						e("div", "id", "editor").R(t("")),
						e("textarea", "id", "note_body", "class", "note-body", "name", "note_body", "rows", "1").
							R(t(note.Body)),
					),
					e("div", "class", "action-btns").R(
						e("p").R(
							e("input", "type", "submit", "class", "action-btn", "value", "Cancel", "formaction", "/"),
							func() (r int) {
								if note.Id > 0 {
									e("input", "type", "submit", "class", "action-btn dup", "value", "Dup", "formaction", "/dup/"+strNoteId)
									// e("button", "onclick", "javascript:window.location='/duplicate/"+strNoteId+"'").R(t("Duplicate"))
								}
								return
							}(),
							e("input", "type", "submit", "id", "create_update_btn", "class", "action-btn",
								"onsubmit", "getEditorContents", "value", formAction),
						),
					),
				),
			),
			e("script", "type", "text/javascript").R(
				t(`
					require.config({ paths: { 'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.0/min/vs' }});
					require(['vs/editor/editor.main'], function() {

						// Make our own theme
						monaco.editor.defineTheme('ro-dark', {
							base: 'vs-dark',
							inherit: true,
							rules: [
								{ background: '1d1f21' },
								{ token: 'comment', foreground: '909090' },
								{ token: 'string', foreground: 'b5bd68' },
								{ token: 'variable', foreground: 'c5c8c6' },
								{ token: 'keyword', foreground: 'ba7d57' },
								{ token: 'number', foreground: 'de935f' },
							],
							colors: {
								'editorBackground': '#1d1f21',
								// 'editorForeground': '#c5c8c6',
								// 'editor.selectionBackground': '#373b41',
								'editorCursor.foreground': '#6DDADA',
								'editor.lineHighlightBackground': '#606060',
							}
						});

						var init_val = document.getElementById("note_body").value;
						var editor = monaco.editor.create(document.getElementById('editor'), {
							value: init_val,
							language: 'markdown',
							theme: 'ro-dark',
							lineNumbers: 'on',
							minimap: {
								enabled: false
							},
							renderLineHighlight: 'gutter'
						});
						
						var nf = document.getElementById("note_form");
						nf.addEventListener("submit", function() {
							var nb = document.getElementById("note_body");
							if (typeof editor !== 'undefined' && nb !== null && typeof nb !== 'undefined') {
								nb.value = editor.getValue();
							}
							return true;
						});

						// Optionally add a custom keyboard shortcut
						editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyP, function() {
							editor.trigger('keyboard', 'editor.action.quickCommand');
						});
					});
				`),
			),
		),
	)

	_, err = fmt.Fprint(w, s.String())
	if err != nil {
		log.Println("Error on NoteForm render:", err)
	}

	return
}
