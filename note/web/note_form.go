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
	body { background-color: #3a3939; color: #b7b9be }
	.container { padding: 1em; border: 1px solid gray; border-radius: 0.5em;
		width: calc(100vw - 3rem); height: calc(100vh - 5rem);
	}
    #editor { 
        position: relative;
        height: calc(100vh - 17rem);
    }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #515c57}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #515c57; padding: 0.3em 0.3em}
    td label {margin-right: 0.4em; font-size: 0.9em; color: #858181 }
    p label {margin-right: 0.4em; vertical-align: top; font-size: 0.9em; color: #858181}
    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em 4em }
    .h1 { font-size: 1.2em; margin-right: 0.2em; margin-bottom: 0.1em; padding: 0.1em }
	.h1 a {text-decoration:none}
	.h1 a:visited, .h1 a:link {color:7bb197}
    .h3 { color:#b4b4b4; font-size: 0.9rem; font-weight:bold; margin-bottom: 0.1em;
		padding: 0.1em;  font-size: 0.9rem;}
    .title { font-size:1.1em; font-weight: bold; color:green; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#401020; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }
	input.descr { width:99%; background-color:#a29b90; }
	#note_form { width: 100%; height: 100% }
    .note-body { padding-left:1.5em; margin-top: 0.1em; width:99%}
	button {cursor: pointer; margin: 0.5em 0.1em; vertical-align: baseline;}
	td input { background-color:tan; margin-right: 0.8em; width:96% }
	.action-btns { text-align: right }
	input.action-btn { width: 10em; padding-left: 0.2em; padding-right: 0.2em;
		margin-right: 2em; background-color:#a29b90; }
	input.action-btn.dup { width: 6em }
	textarea.note-body { display:none }`)),
		),
		e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.12/ace.min.js").R(),
		e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.12/mode-markdown.min.js").R(),
		e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.12/theme-twilight.min.js").R(),
		// e("script", "type", "text/javascript", "src", "https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.12/theme-solarized_dark.min.js").R(),
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
						e("label", "for", "note_body").R(t("Body"), e("br")),
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
				t(`var editor = ace.edit("editor");
					editor.setTheme("ace/theme/twilight");
					editor.session.setMode("ace/mode/markdown");
					editor.session.setUseWorker(false);
					console.log("JavaScript loaded");
					
					document.getElementById('editor').style.fontSize='15px';
					editor.getSession().setValue(document.getElementById("note_body").value);

					var nf = document.getElementById("note_form");
					nf.addEventListener("submit", getEditorContents);
					function getEditorContents() {
						var nb = document.getElementById("note_body");
						if (typeof editor !== 'undefined' && nb !== null && typeof nb !== 'undefined') {
							nb.value = editor.getValue();
						}
						return true;
					}
				`),
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
