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

	b, e, t := element.Vars()
	e("html").R(
		e("head").R(
			e("title").R(t("GoNotes List")),
			e("style").R(t(string(noteListStyles))),
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
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/new").R(t("New")),
				t(" | "),
				e("a", "class", "tool", "href", "http://127.0.0.1:"+optsStr["port"]+"/q/all").R(t("All")),
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

			e("script", "type", "text/javascript").R(t(`
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
			`)),

			e("script", "type", "text/javascript").R(t(
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
			)),
		),
	)
	_, err = fmt.Fprint(w, b.String())
	return
}
