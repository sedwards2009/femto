package femto

import (
	"regexp"
	"strings"
)

func (v *View) searchDown(r *regexp.Regexp, start, end Loc) bool {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		var l []byte
		var charPos int
		if i == start.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if start.X >= len(runes) {
				start.X = len(runes) - 1
			}
			if start.X < 0 {
				start.X = 0
			}
			l = []byte(string(runes[start.X:]))
			charPos = start.X

			if strings.Contains(r.String(), "^") && start.X != 0 {
				continue
			}
		} else {
			l = v.Buf.lines[i].data
		}

		match := r.FindIndex(l)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{charPos + runePos(match[0], string(l)), i})
			v.Cursor.SetSelectionEnd(Loc{charPos + runePos(match[1], string(l)), i})
			v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
			v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
			v.Cursor.Loc = v.Cursor.CurSelection[1]

			return true
		}
	}
	return false
}

func (v *View) searchUp(r *regexp.Regexp, start, end Loc) bool {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i >= end.Y; i-- {
		var l []byte
		if i == start.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if start.X >= len(runes) {
				start.X = len(runes) - 1
			}
			if start.X < 0 {
				start.X = 0
			}
			l = []byte(string(runes[:start.X]))

			if strings.Contains(r.String(), "$") && start.X != Count(string(l)) {
				continue
			}
		} else {
			l = v.Buf.lines[i].data
		}

		match := r.FindIndex(l)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{runePos(match[0], string(l)), i})
			v.Cursor.SetSelectionEnd(Loc{runePos(match[1], string(l)), i})
			v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
			v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
			v.Cursor.Loc = v.Cursor.CurSelection[1]

			return true
		}
	}
	return false
}

// Search searches in the view for the given regex. The down bool
// specifies whether it should search down from the searchStart position
// or up from there
func (v *View) Search(searchStr string, down bool) {
	if searchStr == "" {
		return
	}
	r, err := regexp.Compile(searchStr)
	if v.Buf.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + searchStr)
	}
	if err != nil {
		return
	}

	spawner := v.Buf.cursors[len(v.Buf.cursors)-1]
	searchStart := spawner.CurSelection[1]

	var found bool
	if down {
		found = v.searchDown(r, searchStart, v.Buf.End())
		if !found {
			found = v.searchDown(r, v.Buf.Start(), searchStart)
		}
	} else {
		found = v.searchUp(r, searchStart, v.Buf.Start())
		if !found {
			found = v.searchUp(r, v.Buf.End(), searchStart)
		}
	}
	if !found {
		v.Cursor.ResetSelection()
	}
}
