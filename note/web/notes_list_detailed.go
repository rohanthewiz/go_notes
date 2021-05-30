package web

import (
	"fmt"
	"go_notes/note"
	"html"
	"io"
	"strconv"
	"strings"

	"github.com/rohanthewiz/element"
	blackfriday "github.com/rohanthewiz/go_markdown"
)

func NotesListDetailed(w io.Writer, notes []note.Note, optsStr map[string]string) (err error) {
	notesCount := len(notes)

	s := &strings.Builder{}
	e := func(el string, p ...string) element.Element {
		return element.New(s, el, p...)
	}
	t := func(p ...string) element.Element {
		return element.Text(s, p...)
	}

	e("html").R(
		e("head").R(
			e("title").R(t("Detailed Notes List")),
			e("style").R(t(`
body { background-color: tan }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #531C1C}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #8A2E2E; padding: 0.3em 0.3em}
    li { border-top: 1px solid #B89c72; line-height:1.2em; padding: 1.2em, 4em }
    .h1 { font-size: 1.2em; margin-bottom: 0.1em; padding: 0.1em }
	.h1 a {text-decoration:none}
	.h1 a:visited {color:black}
	.li a {text-decoration:none}
	.li a:visited {color:black}
    .title { font-weight: bold; color:darkgreen; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#401020; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#401020; padding-left: 0.5em }
    .note-body { padding-left:1em; margin-top: 0.1em}
    code { -webkit-border-radius: 0.3em;
          -moz-border-radius: 0.3em;
          border-radius: 0.3em; }			
			`)),
			e("link", "rel", "stylesheet", "href",
				"//cdnjs.cloudflare.com/ajax/libs/highlight.js/8.4/styles/zenburn.min.css").R(),
			e("script", "type", "text/javascript", "src",
				"https://code.jquery.com/jquery-2.1.3.min.js").R(),
			e("script", "type", "text/javascript", "src",
				"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/8.4/highlight.min.js").R(),
		),
		e("body").R(
			e("p").R(
				e("span", "class", "h1").R(
					e("a", "href", "/").R(t("GoNotes")),
				),
				e("span", "class", "count").R(
					t(strconv.Itoa(notesCount), " found"),
				),
				t("["),
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/new").R(t("New")),
				t(" | "),
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/q/all").R(t("All")),
				t("]"),
			),
			e("ul", "class", "topmost").R(
				func() (out element.Element) {
					for _, n := range notes {
						strId := strconv.FormatUint(n.Id, 10)
						e("li").R(
							e("a", "class", "title", "href", "/show/"+strId).R(t(html.EscapeString(n.Title))),
							t(" "),
							e("a", "href", "/edit/"+strId).R(t("edit")),
							t(" | "),
							e("a", "href", "/del/"+strId,
								"onclick", "return confirm('Are you sure you want to delete this note?')",
							).R(t("delete")),
							func() (o element.Element) {
								if n.Description != "" {
									t(" ", html.EscapeString(n.Description))
								}
								return
							}(),
							func() (o element.Element) {
								if n.Body != "" {
									o = e("div", "class", "note-body").R(
										t(string(blackfriday.MarkdownCommon([]byte(n.Body)))),
									)
								}
								return
							}(),
						)
					}
					return out
				}(),
			),
			e("script", "type", "text/javascript").R(t(
				`$( function() {el = $('.note-body');
				el.find("pre code").each( function(i, block) {
					hljs.highlightBlock( block ); }) });`,
			)),
		),
	)
	_, err = fmt.Fprint(w, s.String())
	// fmt.Println(s.String())

	return
}
