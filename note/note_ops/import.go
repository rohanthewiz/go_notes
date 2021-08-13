package note_ops

import (
	"go_notes/config"
	"strings"
)

func HandleImport() {
	o := config.Opts

	arr := strings.Split(o.Import, ".")
	last := len(arr) - 1
	if arr[last] == "csv" {
		ImportCsv(o.Import)
	}
	if arr[last] == "gob" {
		ImportGob(o.Import)
	}
}
