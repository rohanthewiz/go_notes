package note

import (
	"fmt"
	"go_notes/config"
)

func ListNotes(notes []Note, showCount bool) {
	fmt.Println(lineSeparator)
	for _, n := range notes {
		fmt.Printf("[%d] %s", n.Id, n.Title)
		if n.Description != "" {
			fmt.Printf(" - %s", n.Description)
		}
		fmt.Println("")
		if !config.Opts.Short {
			if n.Body != "" {
				fmt.Println(n.Body)
			}
			if n.Tag != "" {
				fmt.Println("Tags:", n.Tag)
			}
		}
		fmt.Println(lineSeparator)
	}
	if showCount {
		var msg string
		if len(notes) != 1 {
			msg = "s"
		}
		fmt.Printf("(%d note%s found)\n", len(notes), msg)
	}
}
