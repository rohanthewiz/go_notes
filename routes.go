package main

import (
	"github.com/julienschmidt/httprouter"
)

func doRoutes(router * httprouter.Router) {
	router.GET("/", Index)
	router.GET("/q/:query", Query)
	router.GET("/new", WebNoteForm)
	router.POST("/create", WebCreateNote)
}
