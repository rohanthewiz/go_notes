// Generated by ego on Sun May 31 21:47:37 2015.
// DO NOT EDIT

package main

import (
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"strconv"
)

//line templates/NoteForm.ego:1
func RenderNoteForm(w io.Writer, note note.Note) error {
//line templates/NoteForm.ego:2
	_, _ = fmt.Fprint(w, "\n")
//line templates/NoteForm.ego:3
	_, _ = fmt.Fprint(w, "\n")
//line templates/NoteForm.ego:3
	var action, button string
	if len(note.Title) > 0 {
		action = "/note/" + strconv.FormatUint(note.Id, 10)
		button = "Update"
	} else {
		action = "/create/"
		button = "Create"
	}

//line templates/NoteForm.ego:12
	_, _ = fmt.Fprint(w, "\n<html>\n<head>\n  <style>\n    body { background-color: tan }\n    h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }\n    .container { padding: 1em; border: 1px solid gray; border-radius: 0.5em }\n    .title { font-weight: bold; color:darkgreen }\n    .note-body { padding-left:1.5em;}\n    input { background-color: #EEE6D0 }\n    textarea { background-color: #ECE6D0 }\n  </style>\n\n</head>\n\n<body>\n<h1>Note</h1>\n\n<div class=\"container\">\n  <form action=\"")
//line templates/NoteForm.ego:30
	_, _ = fmt.Fprint(w, html.EscapeString(fmt.Sprintf("%v", action)))
//line templates/NoteForm.ego:30
	_, _ = fmt.Fprint(w, "\" method=\"post\">\n    <table><tr>\n      <td>\n        <label for=\"title\">Title</label>\n        <input name=\"title\" type = \"text\" value=\"")
//line templates/NoteForm.ego:34
	_, _ = fmt.Fprint(w, html.EscapeString(fmt.Sprintf("%v", note.Title)))
//line templates/NoteForm.ego:34
	_, _ = fmt.Fprint(w, "\" size=54 />\n      </td><td>\n        <label for=\"tag\">&nbsp;&nbsp;Tags</label>\n        <input name=\"tag\" type = \"text\" value=\"")
//line templates/NoteForm.ego:37
	_, _ = fmt.Fprint(w, html.EscapeString(fmt.Sprintf("%v", note.Tag)))
//line templates/NoteForm.ego:37
	_, _ = fmt.Fprint(w, "\" size=24 />\n      <td></tr>\n    </table>\n    <p>\n      <label for=\"description\">Description</label>\n      <input name=\"description\" type = \"text\" value=\"")
//line templates/NoteForm.ego:42
	_, _ = fmt.Fprint(w, html.EscapeString(fmt.Sprintf("%v", note.Description)))
//line templates/NoteForm.ego:42
	_, _ = fmt.Fprint(w, "\" size=83/>\n    </p>\n    <p>\n      <label for=\"body\">Body</label><br>\n      <textarea name=\"body\" rows=\"14\" cols=\"76\">")
//line templates/NoteForm.ego:46
	_, _ = fmt.Fprintf(w, "%v", note.Body)
//line templates/NoteForm.ego:46
	_, _ = fmt.Fprint(w, "</textarea>\n    </p>\n    <p>\n      <input type=\"submit\" value = \"")
//line templates/NoteForm.ego:49
	_, _ = fmt.Fprint(w, html.EscapeString(fmt.Sprintf("%v", button)))
//line templates/NoteForm.ego:49
	_, _ = fmt.Fprint(w, "\" />\n    </p>\n  </form>\n</div>\n\n</body>\n</html>\n")
	return nil
}

//line templates/query.ego:1
