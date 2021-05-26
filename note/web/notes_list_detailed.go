package web

import (
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"strconv"

	"github.com/rohanthewiz/element"
	blackfriday "github.com/rohanthewiz/go_markdown"
)

func NotesListDetailed(w io.Writer, notes []note.Note, optsStr map[string]string) (err error) {
	notesCount := len(notes)

	e := element.New
	str := e("html").R(
		e("head").R(
			e("style").R(`
body { background-color: tan }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #531C1C}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #8A2E2E; padding: 0.3em 0.3em}
    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em, 4em }
    .h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }
    .title { font-weight: bold; color:darkgreen; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#401020; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }
    .note-body { padding-left:1em; margin-top: 0.1em}
    code { -webkit-border-radius: 0.3em;
          -moz-border-radius: 0.3em;
          border-radius: 0.3em; }			
			`),
			e("link", "rel", "stylesheet", "href",
				"//cdnjs.cloudflare.com/ajax/libs/highlight.js/8.4/styles/zenburn.min.css").R(),
			e("script", "type", "text/javascript", "src",
				"https://code.jquery.com/jquery-2.1.3.min.js").R(),
			e("script", "type", "text/javascript", "src",
				"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/8.4/highlight.min.js").R(),
		),
		e("body").R(
			e("p").R(
				`<span class="h1">GoNotes</span>`,
				e("span", "class", "count").R(
					strconv.Itoa(notesCount), " found",
				),
				"[",
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/new").R("New"),
				" | ",
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/q/all").R("All"),
				"]",
			),
			e("ul", "class", "topmost").R(
				func() (out string) {
					for _, n := range notes {
						strId := strconv.FormatUint(n.Id, 10)
						out += e("li").R(
							e("a", "class", "title", "href", "/show/"+strId).R(html.EscapeString(n.Title)),
							" ",
							e("a", "class", "title", "href", "/edit/"+strId).R("edit"),
							" | ",
							e("a", "class", "title", "href", "/del/"+strId,
								"onclick", "return confirm('Are you sure you want to delete this note?')",
							).R("delete"),
							func() (o string) {
								if n.Description != "" {
									o = " " + html.EscapeString(n.Description)
								}
								return
							}(),
							func() (o string) {
								if n.Body != "" {
									o = e("div", "class", "note-body").R(
										string(blackfriday.MarkdownCommon([]byte(n.Body))),
									)
								}
								return
							}(),
						)
					}
					return out
				}(),
			),
			e("script", "type", "text/javascript").R(
				`$( function() {el = $('.note-body');
				el.find("pre code").each( function(i, block) {
					hljs.highlightBlock( block ); }) });`,
			),
		),
	)
	_, err = fmt.Fprint(w, str)
	// fmt.Println(str)

	return
}
