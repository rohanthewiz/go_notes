package web

import (
	"github.com/gofiber/fiber/v2"
)

func DoRoutes(router *fiber.App) {
	router.Get("/", Index)
	router.Post("/", Index)
	router.Get("/ql", QueryLast)
	router.Get("/qi/:id", QueryId)
	router.Get("/json/qi/:id", QueryIdAsJson)
	router.Get("/qg/:tag", QueryTag)
	router.Get("/g/:tag", QueryTag)
	router.Get("/qt/:title", QueryTitle)
	router.Get("/q/:query", Query)
	router.Get("/q/:query/l/:limit", Query)
	router.Get("/qg/:tag/q/:query", QueryTagAndWildCard)
	router.Get("/q/:query/qg/:tag", QueryTagAndWildCard)
	router.Get("/g/:tag/:query", QueryTagAndWildCard)
	router.Get("/qt/:title/q/:query", QueryTitleAndWildCard)
	router.Get("/q/:query/qt/:title", QueryTitleAndWildCard)
	router.Get("/show/:id", QueryId)
	router.Get("/new", WebNoteForm)
	router.Get("/edit/:id", WebNoteForm)
	router.Get("/del/:id", WebDeleteNote)
	router.Get("/js/:file", ServeJS)
	router.Post("/create", WebCreateNote)
	router.Post("/note/:id", WebUpdateNote)
	router.Post("/dup/:id", WebNoteDup)
	router.Get("/dup/:id", WebNoteDup)
}
