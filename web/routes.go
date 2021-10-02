package web

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func DoRoutes(app *fiber.App) {
	app.Static("/assets", "./dist", fiber.Static{
		Compress:      true,
		Browse:        false,
		CacheDuration: 1 * time.Minute,
	})

	app.Get("/", Index)
	app.Post("/", Index)
	app.Get("/ql", QueryLast)
	app.Get("/qi/:id", QueryId)
	app.Get("/json/qi/:id", QueryIdAsJson)
	app.Get("/qg/:tag", QueryTag)
	app.Get("/g/:tag", QueryTag)
	app.Get("/qt/:title", QueryTitle)
	app.Get("/q/:query", Query)
	app.Get("/q/:query/l/:limit", Query)
	app.Get("/qg/:tag/q/:query", QueryTagAndWildCard)
	app.Get("/q/:query/qg/:tag", QueryTagAndWildCard)
	app.Get("/g/:tag/:query", QueryTagAndWildCard)
	app.Get("/qt/:title/q/:query", QueryTitleAndWildCard)
	app.Get("/q/:query/qt/:title", QueryTitleAndWildCard)
	app.Get("/show/:id", QueryId)
	app.Get("/new", WebNoteForm)
	app.Get("/edit/:id", WebNoteForm)
	app.Get("/del/:id", WebDeleteNote)
	// app.Get("/js/:file", ServeJS)
	app.Post("/create", WebCreateNote)
	app.Post("/note/:id", WebUpdateNote)
	app.Post("/dup/:id", WebNoteDup)
	app.Get("/dup/:id", WebNoteDup)
}
