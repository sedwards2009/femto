package femto

import (
	"regexp"
	"unicode/utf8"

	"github.com/pgavlin/femto/util"
)

/*
func (v *View) searchDown(r *regexp.Regexp, start Loc, end Loc) bool {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		var lineBytes []byte
		var charPos int
		if i == start.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if start.X >= len(runes) {
				start.X = len(runes) - 1
			}
			if start.X < 0 {
				start.X = 0
			}
			lineBytes = []byte(string(runes[start.X:]))
			charPos = start.X

			if strings.Contains(r.String(), "^") && start.X != 0 {
				continue
			}
		} else {
			lineBytes = v.Buf.lines[i].data
		}

		match := r.FindIndex(lineBytes)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{charPos + runePos(match[0], string(lineBytes)), i})
			v.Cursor.SetSelectionEnd(Loc{charPos + runePos(match[1], string(lineBytes)), i})
			v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
			v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
			v.Cursor.Loc = v.Cursor.CurSelection[1]

			return true
		}
	}
	return false
}

func (v *View) searchUp(r *regexp.Regexp, start Loc, end Loc) bool {
	currentLoc := start
	if currentLoc.Y >= v.Buf.NumLines {
		currentLoc.Y = v.Buf.NumLines - 1
	}
	if currentLoc.Y < 0 {
		currentLoc.Y = 0
	}
	for i := currentLoc.Y; i >= end.Y; i-- {
		var lineBytes []byte
		if i == currentLoc.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if currentLoc.X >= len(runes) {
				currentLoc.X = len(runes) - 1
			}
			if currentLoc.X < 0 {
				currentLoc.X = 0
			}
			lineBytes = []byte(string(runes[:currentLoc.X]))

			if strings.Contains(r.String(), "$") && currentLoc.X != Count(string(lineBytes)) {
				continue
			}
		} else {
			lineBytes = v.Buf.lines[i].data
		}

		match := r.FindIndex(lineBytes)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{runePos(match[0], string(lineBytes)), i})
			v.Cursor.SetSelectionEnd(Loc{runePos(match[1], string(lineBytes)), i})
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
*/

// We want "^" and "$" to match only the beginning/end of a line, not the
// beginning/end of the search region if it is in the middle of a line.
// In that case we use padded regexps to require a rune before or after
// the match. (This also affects other empty-string patters like "\\b".)
// The following two flags indicate the padding used.
const (
	padStart = 1 << iota
	padEnd
)

func findLineParams(b *Buffer, start, end Loc, i int, r *regexp.Regexp) ([]byte, int, int, *regexp.Regexp) {
	l := b.LineBytes(i)
	charpos := 0
	padMode := 0

	if i == end.Y {
		nchars := util.CharacterCount(l)
		end.X = util.Clamp(end.X, 0, nchars)
		if end.X < nchars {
			l = util.SliceStart(l, end.X+1)
			padMode |= padEnd
		}
	}

	if i == start.Y {
		nchars := util.CharacterCount(l)
		start.X = util.Clamp(start.X, 0, nchars)
		if start.X > 0 {
			charpos = start.X - 1
			l = util.SliceEnd(l, charpos)
			padMode |= padStart
		}
	}

	if padMode == padStart {
		r = regexp.MustCompile(".(?:" + r.String() + ")")
	} else if padMode == padEnd {
		r = regexp.MustCompile("(?:" + r.String() + ").")
	} else if padMode == padStart|padEnd {
		r = regexp.MustCompile(".(?:" + r.String() + ").")
	}

	return l, charpos, padMode, r
}

func (b *Buffer) findDown(r *regexp.Regexp, start, end Loc) ([2]Loc, bool) {
	lastcn := util.CharacterCount(b.LineBytes(b.LinesNum() - 1))
	if start.Y > b.LinesNum()-1 {
		start.X = lastcn - 1
	}
	if end.Y > b.LinesNum()-1 {
		end.X = lastcn
	}
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := start.Y; i <= end.Y; i++ {
		l, charpos, padMode, rPadded := findLineParams(b, start, end, i, r)

		match := rPadded.FindIndex(l)

		if match != nil {
			if padMode&padStart != 0 {
				_, size := utf8.DecodeRune(l[match[0]:])
				match[0] += size
			}
			if padMode&padEnd != 0 {
				_, size := utf8.DecodeLastRune(l[:match[1]])
				match[1] -= size
			}
			start := Loc{charpos + util.RunePos(l, match[0]), i}
			end := Loc{charpos + util.RunePos(l, match[1]), i}
			return [2]Loc{start, end}, true
		}
	}
	return [2]Loc{}, false
}

func (b *Buffer) findUp(r *regexp.Regexp, start, end Loc) ([2]Loc, bool) {
	lastcn := util.CharacterCount(b.LineBytes(b.LinesNum() - 1))
	if start.Y > b.LinesNum()-1 {
		start.X = lastcn - 1
	}
	if end.Y > b.LinesNum()-1 {
		end.X = lastcn
	}
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := end.Y; i >= start.Y; i-- {
		charCount := util.CharacterCount(b.LineBytes(i))
		from := Loc{0, i}.Clamp(start, end)
		to := Loc{charCount, i}.Clamp(start, end)

		allMatches := b.findAll(r, from, to)
		if allMatches != nil {
			match := allMatches[len(allMatches)-1]
			return [2]Loc{match[0], match[1]}, true
		}
	}
	return [2]Loc{}, false
}

func (b *Buffer) findAll(r *regexp.Regexp, start, end Loc) [][2]Loc {
	var matches [][2]Loc
	loc := start
	for {
		match, found := b.findDown(r, loc, end)
		if !found {
			break
		}
		matches = append(matches, match)
		if match[0] != match[1] {
			loc = match[1]
		} else if match[1] != end {
			loc = match[1].Move(1, b)
		} else {
			break
		}
	}
	return matches
}

// FindNext finds the next occurrence of a given string in the buffer
// It returns the start and end location of the match (if found) and
// a boolean indicating if it was found
// May also return an error if the search regex is invalid
func (b *Buffer) FindNext(s string, start, end, from Loc, down bool, useRegex bool) ([2]Loc, bool, error) {
	if s == "" {
		return [2]Loc{}, false, nil
	}

	var r *regexp.Regexp
	var err error

	if !useRegex {
		s = regexp.QuoteMeta(s)
	}

	if b.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + s)
	} else {
		r, err = regexp.Compile(s)
	}

	if err != nil {
		return [2]Loc{}, false, err
	}

	var found bool
	var l [2]Loc
	if down {
		l, found = b.findDown(r, from, end)
		if !found {
			l, found = b.findDown(r, start, end)
		}
	} else {
		l, found = b.findUp(r, from, start)
		if !found {
			l, found = b.findUp(r, end, start)
		}
	}
	return l, found, nil
}

// ReplaceRegex replaces all occurrences of 'search' with 'replace' in the given area
// and returns the number of replacements made and the number of characters
// added or removed on the last line of the range
func (b *Buffer) ReplaceRegex(start, end Loc, search *regexp.Regexp, replace []byte, captureGroups bool) (int, int) {
	if start.GreaterThan(end) {
		start, end = end, start
	}

	charsEnd := util.CharacterCount(b.LineBytes(end.Y))
	found := 0
	var deltas []Delta

	for i := start.Y; i <= end.Y; i++ {
		l := b.LineBytes(i)
		charCount := util.CharacterCount(l)
		if (i == start.Y && start.X > 0) || (i == end.Y && end.X < charCount) {
			// This replacement code works in general, but it creates a separate
			// modification for each match. We only use it for the first and last
			// lines, which may use padded regexps

			from := Loc{0, i}.Clamp(start, end)
			to := Loc{charCount, i}.Clamp(start, end)
			matches := b.findAll(search, from, to)
			found += len(matches)

			for j := len(matches) - 1; j >= 0; j-- {
				// if we counted upwards, the different deltas would interfere
				match := matches[j]
				var newText []byte
				if captureGroups {
					newText = search.ReplaceAll(b.Substr(match[0], match[1]), replace)
				} else {
					newText = replace
				}
				deltas = append(deltas, Delta{newText, match[0], match[1]})
			}
		} else {
			newLine := search.ReplaceAllFunc(l, func(in []byte) []byte {
				found++
				var result []byte
				if captureGroups {
					match := search.FindSubmatchIndex(in)
					result = search.Expand(result, replace, in, match)
				} else {
					result = replace
				}
				return result
			})
			deltas = append(deltas, Delta{newLine, Loc{0, i}, Loc{charCount, i}})
		}
	}

	b.MultipleReplace(deltas)

	return found, util.CharacterCount(b.LineBytes(end.Y)) - charsEnd
}

// Search searches for a given string/regex in the buffer and selects the next
// match if a match is found
// This function behaves the same way as Find and FindLiteral actions:
// it affects the buffer's LastSearch and LastSearchRegex (saved searches)
// for use with FindNext and FindPrevious, and turns HighlightSearch on or off
// according to hlsearch setting
func (v *View) Search(str string, useRegex bool, searchDown bool) error {
	loc := v.Cursor.Loc
	if v.Cursor.HasSelection() && !searchDown {
		loc = v.Cursor.CurSelection[0]
	}

	match, found, err := v.Buf.FindNext(str, v.Buf.Start(), v.Buf.End(), loc, searchDown, useRegex)
	if err != nil {
		return err
	}

	if found {
		v.Cursor.SetSelectionStart(match[0])
		v.Cursor.SetSelectionEnd(match[1])
		v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
		v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
		v.Cursor.GotoLoc(v.Cursor.CurSelection[1])
	} else {
		v.Cursor.ResetSelection()
	}
	return nil
}
