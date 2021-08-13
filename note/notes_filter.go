package note

import (
	"go_notes/config"
	"strings"
)

type NotesFilter struct {
	Id       int64 // qi
	Title    string
	Tags     []string
	QueryStr string
	Last     bool
	Limit    int
	Offset   int
}

func NotesFilterFromOpts() (nf *NotesFilter) {
	o := config.Opts // alias
	nf = &NotesFilter{
		Id:       o.QI,
		Tags:     strings.Split(o.Tag, ","),
		Title:    o.Title,
		QueryStr: o.Q,
		Last:     o.Last,
		Limit:    o.Limit,
	}
	return
}
