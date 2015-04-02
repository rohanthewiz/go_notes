package main
import (
"fmt"
"io"
"strconv"
"github.com/russross/blackfriday"
)
//line NoteForm.ego:1
 func RenderNoteForm(w io.Writer, note Note) error  {
//line NoteForm.ego:2
_, _ = fmt.Fprintf(w, "\n")
//line NoteForm.ego:3
_, _ = fmt.Fprintf(w, "\n")
//line NoteForm.ego:3
  var action, button string
    if len(note.Title) > 0 {
      action = "/note/" + strconv.FormatInt(note.Id, 10)
      button = "Update"
    } else {
      action = "/create/"
      button = "Create"
    }

//line NoteForm.ego:12
_, _ = fmt.Fprintf(w, "\n<html>\n<head>\n  <style>\n    body { background-color: tan }\n    h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }\n    .container { padding: 1em; border: 1px solid gray; border-radius: 0.5em }\n    .title { font-weight: bold; color:darkgreen }\n    .note-body { padding-left:1.5em;}\n    input { background-color: #EEE6D0 }\n    textarea { background-color: #ECE6D0 }\n  </style>\n\n  <script type=\"text/javascript\" src=\"js/funcs.js\">\n  </script>\n</head>\n\n<body>\n<h1>Note</h1>\n\n<div class=\"container\">\n  <form action=\"")
//line NoteForm.ego:32
_, _ = fmt.Fprintf(w, "%v",  action )
//line NoteForm.ego:32
_, _ = fmt.Fprintf(w, "\" method=\"post\">\n    <p>\n      <label for=\"title\">Title</label>\n      <input name=\"title\" type = \"text\" value=\"")
//line NoteForm.ego:35
_, _ = fmt.Fprintf(w, "%v",  note.Title )
//line NoteForm.ego:35
_, _ = fmt.Fprintf(w, "\" size=54 />\n    </p>\n    <p>\n      <label for=\"description\">Description</label>\n      <input name=\"description\" type = \"text\" value=\"")
//line NoteForm.ego:39
_, _ = fmt.Fprintf(w, "%v",  note.Description )
//line NoteForm.ego:39
_, _ = fmt.Fprintf(w, "\" size=60 />\n    </p>\n    <p>\n      <label for=\"body\">Body</label><br>\n      <textarea name=\"body\" rows=\"14\" cols=\"76\">")
//line NoteForm.ego:43
_, _ = fmt.Fprintf(w, "%v",  note.Body )
//line NoteForm.ego:43
_, _ = fmt.Fprintf(w, "</textarea>\n    </p>\n    <p>\n      <label for=\"tag\">Tags</label>\n      <input name=\"tag\" type = \"text\" value=\"")
//line NoteForm.ego:47
_, _ = fmt.Fprintf(w, "%v",  note.Tag )
//line NoteForm.ego:47
_, _ = fmt.Fprintf(w, "\" />\n    </p>\n    <p>\n      <input type=\"submit\" value = \"")
//line NoteForm.ego:50
_, _ = fmt.Fprintf(w, "%v",  button )
//line NoteForm.ego:50
_, _ = fmt.Fprintf(w, "\" />\n    </p>\n  </form>\n</div>\n<!-- <script>\n  doAlert(\"Hi there from Javascript!\\nFunny chars test: %% is a percentage\");\n</script>\n-->\n\n</body>\n</html>\n")
return nil
}
//line query.ego:1
 func RenderQuery(w io.Writer, notes []Note) error  {
//line query.ego:2
_, _ = fmt.Fprintf(w, "\n")
//line query.ego:3
_, _ = fmt.Fprintf(w, "\n")
//line query.ego:4
_, _ = fmt.Fprintf(w, "\n\n<html>\n<head>\n  <style>\n    body { background-color: tan }\n    ul { list-style-type:none; margin: 0; padding: 0; }\n    ul.topmost > li:first-child { border-top: 1px solid #531C1C}\n    ul.topmost > li { border-top:none; border-bottom: 1px solid #8A2E2E; padding: 0.3em 0.3em}\n    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em, 4em }\n    .h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }\n    .title { font-weight: bold; color:darkgreen; padding-top: 0.4em }\n    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }\n    .note-body { padding-left:1em; margin-top: 0.1em}\n  </style>\n</head>\n<body>\n\n<p>\n  <span class=\"h1\">My Notes</span>  [<a class=\"tool\" href=\"http://127.0.0.1:8080/new\">New</a> |\n  <a class=\"tool\" href=\"http://127.0.0.1:8080/q/all\">All</a>]\n</p>\n\n<ul class=\"topmost\">\n  ")
//line query.ego:27
 for _, note := range notes { 
//line query.ego:28
_, _ = fmt.Fprintf(w, "\n      ")
//line query.ego:28
 id_str := strconv.FormatInt(note.Id, 10) 
//line query.ego:29
_, _ = fmt.Fprintf(w, "\n      <li><a class=\"title\" href=\"http://127.0.0.1:8080/show/")
//line query.ego:29
_, _ = fmt.Fprintf(w, "%v",  id_str )
//line query.ego:29
_, _ = fmt.Fprintf(w, "\">")
//line query.ego:29
_, _ = fmt.Fprintf(w, "%v",  note.Title )
//line query.ego:29
_, _ = fmt.Fprintf(w, "</a>\n      <a class=\"tool\" href=\"http://127.0.0.1:8080/edit/")
//line query.ego:30
_, _ = fmt.Fprintf(w, "%v",  id_str )
//line query.ego:30
_, _ = fmt.Fprintf(w, "\">[edit]</a>")
//line query.ego:30
 if note.Description != "" { 
//line query.ego:30
_, _ = fmt.Fprintf(w, " - ")
//line query.ego:30
_, _ = fmt.Fprintf(w, "%v",  note.Description )
//line query.ego:31
_, _ = fmt.Fprintf(w, "\n      ")
//line query.ego:31
 } 
//line query.ego:32
_, _ = fmt.Fprintf(w, "\n      ")
//line query.ego:32
 if note.Body != "" { 
//line query.ego:33
_, _ = fmt.Fprintf(w, "\n        <br><div class=\"note-body\">")
//line query.ego:33
_, _ = fmt.Fprintf(w, "%v",  string(blackfriday.MarkdownCommon([]byte(note.Body))) )
//line query.ego:33
_, _ = fmt.Fprintf(w, "</span>\n      ")
//line query.ego:34
 } 
//line query.ego:35
_, _ = fmt.Fprintf(w, "\n      </li>\n  ")
//line query.ego:36
 } 
//line query.ego:37
_, _ = fmt.Fprintf(w, "\n</ul>\n\n</body>\n</html>\n")
return nil
}
