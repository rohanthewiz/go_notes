package note

import "strings"

func FilterOutPrivate(notes []Note) (filtered []Note) {
	filtered = make([]Note, 0, len(notes))

	for _, n := range notes {
		tags := strings.Split(n.Tag, ",")
		isPrivate := false

		for _, t := range tags {
			if strings.ToLower(strings.TrimSpace(t)) == "private" {
				isPrivate = true
			}
		}

		if !isPrivate {
			filtered = append(filtered, n)
		}
	}

	return
}
