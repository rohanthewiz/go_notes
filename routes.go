package main

import (
	"github.com/julienschmidt/httprouter"
)

func doRoutes(router * httprouter.Router) {
	router.GET("/", Index)
	router.GET("/qi/:id", QueryId)
	router.GET("/qg/:tag", QueryTag)
	router.GET("/qt/:title", QueryTitle)
	router.GET("/q/:query", Query)
	router.GET("/qg/:tag/q/:query", QueryTagAndWildCard)
	router.GET("/qt/:title/q/:query", QueryTitleAndWildCard)
	router.GET("/show/:id", QueryId)
	router.GET("/new", WebNoteForm)
	router.GET("/edit/:id", WebNoteForm)
	router.GET("/js/:file", ServeJS)
	router.POST("/create", WebCreateNote)
	router.POST("/note/:id", WebUpdateNote)
}
