package web

import (
	_ "embed"
	"fmt"
	"go_notes/note"
	"go_notes/utils"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rohanthewiz/element"
	blackfriday "github.com/rohanthewiz/go_markdown"
)

var (
	//go:embed embeds/notes_list.css
	noteListStyles []byte
)

func NotesList(w io.Writer, req *http.Request, notes []note.Note, optsStr map[string]string) (err error) {
	const NotesDetailsThreshold = 2
	notesCount := len(notes)
	showDetails := notesCount <= NotesDetailsThreshold

	b := element.B()
	b.Html().R(
		b.Head().R(
			b.Title().T("GoNotes List"),
			b.Style().T(string(noteListStyles)),
			b.Link("rel", "stylesheet", "href",
				"//cdnjs.cloudflare.com/ajax/libs/highlight.js/11.2.0/styles/agate.min.css").R(),
			b.Script("type", "text/javascript", "src",
				"https://code.jquery.com/jquery-2.1.3.min.js").R(),
			b.Script("type", "text/javascript", "src",
				"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.2.0/highlight.min.js").R(),
		),
		b.Body().R(
			b.P().R(
				b.SpanClass("h1").R(
					b.A("href", "/").T("GoNotes"),
				),
				b.SpanClass("count").T(strconv.Itoa(notesCount), " found"),
				b.T("["),
				b.AClass("tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/new").T("New"),
				b.T(" | "),
				b.AClass("tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/q/all").T("All"),
				b.T("]"),
				b.Wrap(func() {
					if optsStr["git_commit"] != "" && optsStr["git_commit"] != "dev" {
						b.SpanClass("build-info").T(" | ", optsStr["git_commit"])
					}
				}),
			),
			b.UlClass("topmost").R(
				element.ForEach(notes, func(n note.Note) {
					strId := strconv.FormatUint(n.Id, 10)
					b.Li().R(
						b.AClass("title", "href", "/show/"+strId).T(html.EscapeString(n.Title)),
						b.SpanClass("small").R(
							b.T(" ["),
							b.Wrap(func() {
								localLoc, err := time.LoadLocation("Local")
								if err != nil {
									log.Println(`In NotesList Failed to load location "Local"`)
								} else {
									localDateTime := n.UpdatedAt.In(localLoc)
									b.SpanClass("time-label").T("upd: ", localDateTime.Format("2006-01-02"), " ")
								}
							}),
							b.Wrap(func() {
								if showDetails {
									b.SpanClass("small").T(" GUID ", utils.TruncString(n.Guid, 15))
								}
							}),
							b.T("&nbsp;&nbsp;"),
							b.AClass("text-menu", "href", "/edit/"+strId).T("edit"),
							b.T(" | "),
							b.AClass("text-menu", "href", "/dup/"+strId).T("dup"),
							b.T(" | "),
							b.AClass("text-menu", "href", "/del/"+strId+"?return="+req.URL.Path,
								"onclick", "return confirm('Are you sure you want to delete this note?')").T("del"),
							b.T("] "),
							b.Wrap(func() {
								if showDetails {
									b.Br()
								}
							}),
						),
						b.Wrap(func() {
							if n.Description != "" {
								b.T(" ", html.EscapeString(n.Description))
							}
						}),
						b.Wrap(func() {
							if showDetails && n.Body != "" {
								b.DivClass("note-body").T(string(blackfriday.MarkdownCommon([]byte(n.Body))))
							}
						}),
					)
				}),
			),

			b.Script("type", "text/javascript").T(`
				function copyToClipboard(element) {
						navigator.clipboard.writeText(element.innerText).then(() => {
							// alert('Code copied to clipboard!');
						}).catch(err => {
							console.error('Failed to copy: ', err);
						});
				}

				function addInlineCopyBtn(code) {
							const button = document.createElement('button');
							button.className = 'inline-copy-btn';
							button.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="12" height="11" viewBox="0 0 24 24"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>';
							button.onclick = () => copyToClipboard(code);
							code.insertAdjacentElement('afterend', button);
				}
				function addCopyBtn(pre) {
							const button = document.createElement('button');
							button.className = 'copy-btn';
							button.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>';
							button.onclick = () => copyToClipboard(pre.querySelector('code'));
							pre.parentNode.insertBefore(button, pre);
				}
			`),

			b.Script("type", "text/javascript").T(
				`$( function() {
				el = $('.note-body');

				el.find("pre code").each( function(i, block) {
					hljs.highlightBlock( block );
				});

				el.find("pre").each( function(i, block) {
					addCopyBtn(block);
				});

				el.find("p > code, li > code").each( function(i, block) {
					addInlineCopyBtn( block );
				});
			});`,
			),
		),
	)
	_, err = fmt.Fprint(w, b.String())
	return
}
