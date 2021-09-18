package note

import (
	"bufio"
	"bytes"
	"strings"
)

// Aggregate comments of the format "// ~" into a KeyNote section in the note
func UpsertKeyNotes(nb string) string {
	lnNb := len(nb)
	if lnNb == 0 {
		return nb
	}

	const keyNoteHdrPrefix = "#### Key Notes (auto)"
	const codeSampleLen = 84
	var keyNotes []string
	var inKeyNotes, atKeyNoteHdr bool
	var pastKeyNotes bool
	linesBeforeKeyNote := make([]string, 0, 4) // guesstimates here
	linesAfterKeyNote := make([]string, 0, len(nb)/2)
	sbOut := strings.Builder{}

	scnr := bufio.NewScanner(bytes.NewReader([]byte(nb)))
	for scnr.Scan() { // ~ Scanner splits by default on lines
		line := scnr.Text()
		lineTrimmed := strings.TrimSpace(line)

		// ~ Mark that we are in keynotes
		if strings.HasPrefix(line, keyNoteHdrPrefix) {
			atKeyNoteHdr = true
			atKeyNoteHdr = true
			inKeyNotes = true
			continue
		}

		if lineTrimmed == "" && inKeyNotes && !atKeyNoteHdr { // skip this check immediately after keyNoteHdr
			inKeyNotes = false
			pastKeyNotes = true
		}

		atKeyNoteHdr = false

		// ~ Skip original keynotes
		if inKeyNotes {
			continue
		}

		// ~ Agg lines before keynote
		if !pastKeyNotes {
			linesBeforeKeyNote = append(linesBeforeKeyNote, line)
		} else {
			linesAfterKeyNote = append(linesAfterKeyNote, line)
		}

		// ~ Agg key notes
		tokens := strings.SplitN(line, "// ~", 2)
		if len(tokens) == 2 {
			keyNote := "- " + tokens[1]

			// ~ Fixup any actual code. We add it after the keyNote if not empty
			actualCode := strings.TrimSpace(tokens[0])
			if lnCode := len(actualCode); lnCode > 0 {
				if lnCode > codeSampleLen { // ~ limit the length of code lines
					actualCode = actualCode[:codeSampleLen] + "..."
				}
				keyNote += " `" + actualCode + "`"
			}

			keyNotes = append(keyNotes, keyNote)
		}
	}

	// fmt.Println("keynotes", len(keyNotes), "linesBefore", len(linesBeforeKeyNote),
	// 	"linesAfter", len(linesAfterKeyNote),
	// )

	// ~ Reassemble the note with the keynotes upserted
	// ~ If no keynotes already existed or we got to the end
	// 		write keynotes + linesBefore
	if !pastKeyNotes {
		if len(keyNotes) > 0 {
			sbOut.WriteString(keyNoteHdrPrefix + "\n\n")
			sbOut.WriteString(strings.Join(keyNotes, "\n") + "\n\n")
		}

		if len(linesBeforeKeyNote) > 0 {
			sbOut.WriteString(strings.Join(linesBeforeKeyNote, "\n"))
		}
	} else { // linesBefore + keynote + linesAfter
		if len(linesBeforeKeyNote) > 0 {
			sbOut.WriteString(strings.Join(linesBeforeKeyNote, "\n") + "\n")
			// if len(keyNotes) > 0 {
			// 	sbOut.WriteRune('\n')
			// }
		}
		// ~ Add the keynotes (rem. we discard the original)
		if len(keyNotes) > 0 {
			sbOut.WriteString(keyNoteHdrPrefix + "\n\n")
			sbOut.WriteString(strings.Join(keyNotes, "\n") + "\n")
		}
		if len(linesAfterKeyNote) > 0 {
			sbOut.WriteString(strings.Join(linesAfterKeyNote, "\n"))
		}
	}

	return sbOut.String()
}
