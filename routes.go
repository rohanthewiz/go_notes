package main

import (
	"github.com/julienschmidt/httprouter"
)

func doRoutes(router * httprouter.Router) {
	router.GET("/", Index)
	router.GET("/ql", QueryLast)
	router.GET("/qi/:id", QueryId)
	router.GET("/qg/:tag", QueryTag)
	router.GET("/qt/:title", QueryTitle)
	router.GET("/q/:query", Query)
	router.GET("/q/:query/l/:limit", Query)
	router.GET("/qg/:tag/q/:query", QueryTagAndWildCard)
	router.GET("/q/:query/qg/:tag", QueryTagAndWildCard)
	router.GET("/qt/:title/q/:query", QueryTitleAndWildCard)
	router.GET("/q/:query/qt/:title", QueryTitleAndWildCard)
	router.GET("/show/:id", QueryId)
	router.GET("/new", WebNoteForm)
	router.GET("/edit/:id", WebNoteForm)
	router.GET("/del/:id", WebDeleteNote)
	router.GET("/js/:file", ServeJS)
	router.POST("/create", WebCreateNote)
	router.POST("/note/:id", WebUpdateNote)
}
