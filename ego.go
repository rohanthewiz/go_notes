package main
import (
"fmt"
"io"
)
//line ShowNote.ego:1
 func RenderShowNote(w io.Writer, note Note) error  {
//line ShowNote.ego:2
_, _ = fmt.Fprintf(w, "\n\n<html>\n<head>\n  <style>\n    body { background-color: #faec9a }\n    h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }\n    .container { padding: 1em; border: 1px solid gray; border-radius: 0.5em }\n    .title { font-weight: bold; color:darkgreen }\n    .note-body { padding-left:1.5em;}\n  </style>\n</head>\n<body>\n<h1>Note</h1>\n\n<div class=\"container\">\n  <form action=\"/create\" method=\"post\">\n    <p>\n      <label for=\"title\">Title</label>\n      <input name=\"title\" type = \"text\" value=\"")
//line ShowNote.ego:20
_, _ = fmt.Fprintf(w, "%v",  note.Title )
//line ShowNote.ego:20
_, _ = fmt.Fprintf(w, "\" />\n    </p>\n    <p>\n      <label for=\"description\">Description</label>\n      <input name=\"description\" type = \"text\" value=\"")
//line ShowNote.ego:24
_, _ = fmt.Fprintf(w, "%v",  note.Description )
//line ShowNote.ego:24
_, _ = fmt.Fprintf(w, "\"\n        placeholder=\"Enter description\" />\n    </p>\n    <p>\n      <label for=\"body\">Body</label><br>\n      <textarea name=\"body\" rows=\"4\" cols=\"70\" placeholder=\"Enter note body\">\n        ")
//line ShowNote.ego:30
_, _ = fmt.Fprintf(w, "%v",  note.Body )
//line ShowNote.ego:31
_, _ = fmt.Fprintf(w, "\n      </textarea>\n    </p>\n    <p>\n      <label for=\"tag\">Tags</label>\n      <input name=\"tag\" type = \"text\" value=\"")
//line ShowNote.ego:35
_, _ = fmt.Fprintf(w, "%v",  note.Tag )
//line ShowNote.ego:35
_, _ = fmt.Fprintf(w, "\" />\n    </p>\n    <p>\n      <input type=\"submit\" value = \"create\" />\n    </p>\n  </form>\n</div>\n\n</body>\n</html>\n")
return nil
}
//line query.ego:1
 func RenderQuery(w io.Writer, notes []Note) error  {
//line query.ego:2
_, _ = fmt.Fprintf(w, "\n\n<html>\n<head>\n  <style>\n    body { background-color: #faec9a }\n    ul { list-style-type:none; margin: 0; padding: 0; }\n    li:first-child { border-top: 1px solid #a0a0a0}\n    li { border-bottom: 1px solid #a0a0a0; line-height:1.2em; padding: 1.2em, 4em }\n    h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }\n    .title { font-weight: bold; color:darkgreen }\n    .note-body { padding-left:1.5em;}\n  </style>\n</head>\n<body>\n<h1>My Notes</h1>\n\n<ul>\n  ")
//line query.ego:19
 for _, note := range notes { 
//line query.ego:20
_, _ = fmt.Fprintf(w, "\n          <li><span class=\"title\">")
//line query.ego:20
_, _ = fmt.Fprintf(w, "%v",  note.Title )
//line query.ego:20
_, _ = fmt.Fprintf(w, "</span>\n          ")
//line query.ego:21
 if note.Description != "" { 
//line query.ego:22
_, _ = fmt.Fprintf(w, "\n            - ")
//line query.ego:22
_, _ = fmt.Fprintf(w, "%v",  note.Description )
//line query.ego:23
_, _ = fmt.Fprintf(w, "\n          ")
//line query.ego:23
 } 
//line query.ego:24
_, _ = fmt.Fprintf(w, "\n          ")
//line query.ego:24
 if note.Body != "" { 
//line query.ego:25
_, _ = fmt.Fprintf(w, "\n            <br><span class=\"note-body\">")
//line query.ego:25
_, _ = fmt.Fprintf(w, "%v",  note.Body )
//line query.ego:25
_, _ = fmt.Fprintf(w, "</span>\n          ")
//line query.ego:26
 } 
//line query.ego:27
_, _ = fmt.Fprintf(w, "\n          </li>\n  ")
//line query.ego:28
 } 
//line query.ego:29
_, _ = fmt.Fprintf(w, "\n</ul>\n\n</body>\n</html>\n")
return nil
}
