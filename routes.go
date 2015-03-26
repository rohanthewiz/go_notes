package main

import (
	"github.com/julienschmidt/httprouter"
)

func doRoutes(router * httprouter.Router) {
	router.GET("/", Index)
	router.GET("/q/:query", Query)
	router.GET("/qi/:id", QueryById)
	router.GET("/new", WebNoteForm)
	router.GET("/js/:file", ServeJS)
	router.POST("/create", WebCreateNote)
}
