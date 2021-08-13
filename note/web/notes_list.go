package web

import (
	"fmt"
	"go_notes/config"
	"go_notes/note"
	"go_notes/utils"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rohanthewiz/element"
	blackfriday "github.com/rohanthewiz/go_markdown"
)

func NotesList(w io.Writer, req *http.Request, notes []note.Note) (err error) {
	const NotesDetailsThreshold = 2
	o := config.Opts

	notesCount := len(notes)
	showDetails := notesCount <= NotesDetailsThreshold

	s := &strings.Builder{}
	e := func(el string, p ...string) element.Element {
		return element.New(s, el, p...)
	}
	t := func(p ...string) int {
		return element.Text(s, p...)
	}

	e("html").R(
		e("head").R(
			e("title").R(t("GoNotes List")),
			e("style").R(t(`
body { background-color: #3a3939; color: #b7b9be }
    ul { list-style-type:none; margin: 0; padding: 0; }
    ul.topmost > li:first-child { border-top: 1px solid #515c57}
    ul.topmost > li { border-top:none; border-bottom: 1px solid #515c57; padding: 0.3em 0.3em}
    li { border-top: 1px solid #515c57; line-height:1.2em; padding: 1.2em, 4em }
	li a {text-decoration:none}
	li a:link, li a:visited {color:#acb4b6}
    .h1 { font-size: 1.2em; margin-right: 0.2em; margin-bottom: 0.1em; padding: 0.1em }
	.h1 a {text-decoration:none}
	.h1 a:visited, .h1 a:link {color:#7bb197}
    .h3 { color:#b4b4b4; font-size: 0.9rem; font-weight:bold; margin-bottom: 0.1em;
		padding: 0.1em;  font-size: 0.9rem;}
    .title { font-size:1.1em; font-weight: bold; color:green; padding-top: 0.4em }
    .count { font-size: 0.8em; color:#c4c4c6; padding-left: 0.5em; padding-right: 0.5em }
    .tool { font-size: 0.7em; color:#c6c6c6; padding-left: 0.5em }
    .note-body { padding-left:1em; margin-top: 0.1em}
	.time-label { font-size: 0.7rem }
	.text-menu { font-weight: bold }
	.small { font-size: 0.8em }
    code { border-radius: 0.3em;
    	background-color: #b2916e; color: black;
    	padding: 0.1em 0.3em; }
			`)),
			e("link", "rel", "stylesheet", "href",
				"//cdnjs.cloudflare.com/ajax/libs/highlight.js/11.2.0/styles/agate.min.css").R(),
			e("script", "type", "text/javascript", "src",
				"https://code.jquery.com/jquery-2.1.3.min.js").R(),
			e("script", "type", "text/javascript", "src",
				"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.2.0/highlight.min.js").R(),
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
				e("a", "class", "tool", "href", "http://127.0.0.1:"+o.Port+"/new").R(t("New")),
				t(" | "),
				e("a", "class", "tool", "href", "http://127.0.0.1:"+o.Port+"/q/all").R(t("All")),
				t("]"),
			),
			e("ul", "class", "topmost").R(
				func() (r int) {
					for _, n := range notes {
						strId := strconv.FormatUint(n.Id, 10)
						e("li").R(
							e("a", "class", "title", "href", "/show/"+strId).R(t(html.EscapeString(n.Title))),
							e("span", "class", "small").R(
								t(" ["),
								func() (r int) {
									localLoc, err := time.LoadLocation("Local")
									if err != nil {
										log.Println(`In NotesList Failed to load location "Local"`)
									} else {
										localDateTime := n.UpdatedAt.In(localLoc)
										e("span", "class", "time-label").R(
											t("upd: ", localDateTime.Format("2006-01-02"), " "))
									}
									return
								}(),
								func() (r int) {
									if showDetails {
										e("span", "class", "small").R(t(" GUID ", utils.TruncString(n.Guid, 15)))
									}
									return
								}(),
								t("&nbsp;&nbsp;"),
								e("a", "class", "text-menu", "href", "/edit/"+strId).R(t("edit")),
								t(" | "),
								e("a", "class", "text-menu", "href", "/dup/"+strId).R(t("dup")),
								t(" | "),
								e("a", "class", "text-menu", "href", "/del/"+strId+"?return="+req.URL.Path,
									"onclick", "return confirm('Are you sure you want to delete this note?')",
								).R(t("del")),
								t("] "),
								func() (r int) {
									if showDetails {
										e("br") // single tags don't need an `.R()`
									}
									return
								}(),
							),
							func() (r int) {
								if n.Description != "" {
									t(" ", html.EscapeString(n.Description))
								}
								return
							}(),
							func() (r int) {
								if showDetails && n.Body != "" {
									e("div", "class", "note-body").R(
										t(string(blackfriday.MarkdownCommon([]byte(n.Body)))),
									)
								}
								return
							}(),
						)
					}
					return
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
