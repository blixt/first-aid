package serif

import (
	"strings"
)

type FontStyle int32

const (
	Regular    FontStyle = 0x1D670
	Bold       FontStyle = 0x1D5D4
	BoldItalic FontStyle = 0x1D63C
	Italic     FontStyle = 0x1D608
	// These sets don't look good.
	// Bold       FontStyle = 0x1D400
	// BoldItalic FontStyle = 0x1D468
	// Italic     FontStyle = 0x1D434 // For this one, ℎ is \u210E.
)

type letterMapping struct {
	Offset        int32
	CombiningMark string
}

var letterMap = map[rune]letterMapping{
	'A': {Offset: 0, CombiningMark: ""},
	'B': {Offset: 1, CombiningMark: ""},
	'C': {Offset: 2, CombiningMark: ""},
	'D': {Offset: 3, CombiningMark: ""},
	'E': {Offset: 4, CombiningMark: ""},
	'F': {Offset: 5, CombiningMark: ""},
	'G': {Offset: 6, CombiningMark: ""},
	'H': {Offset: 7, CombiningMark: ""},
	'I': {Offset: 8, CombiningMark: ""},
	'J': {Offset: 9, CombiningMark: ""},
	'K': {Offset: 10, CombiningMark: ""},
	'L': {Offset: 11, CombiningMark: ""},
	'M': {Offset: 12, CombiningMark: ""},
	'N': {Offset: 13, CombiningMark: ""},
	'O': {Offset: 14, CombiningMark: ""},
	'P': {Offset: 15, CombiningMark: ""},
	'Q': {Offset: 16, CombiningMark: ""},
	'R': {Offset: 17, CombiningMark: ""},
	'S': {Offset: 18, CombiningMark: ""},
	'T': {Offset: 19, CombiningMark: ""},
	'U': {Offset: 20, CombiningMark: ""},
	'V': {Offset: 21, CombiningMark: ""},
	'W': {Offset: 22, CombiningMark: ""},
	'X': {Offset: 23, CombiningMark: ""},
	'Y': {Offset: 24, CombiningMark: ""},
	'Z': {Offset: 25, CombiningMark: ""},
	'a': {Offset: 26, CombiningMark: ""},
	'b': {Offset: 27, CombiningMark: ""},
	'c': {Offset: 28, CombiningMark: ""},
	'd': {Offset: 29, CombiningMark: ""},
	'e': {Offset: 30, CombiningMark: ""},
	'f': {Offset: 31, CombiningMark: ""},
	'g': {Offset: 32, CombiningMark: ""},
	'h': {Offset: 33, CombiningMark: ""},
	'i': {Offset: 34, CombiningMark: ""},
	'j': {Offset: 35, CombiningMark: ""},
	'k': {Offset: 36, CombiningMark: ""},
	'l': {Offset: 37, CombiningMark: ""},
	'm': {Offset: 38, CombiningMark: ""},
	'n': {Offset: 39, CombiningMark: ""},
	'o': {Offset: 40, CombiningMark: ""},
	'p': {Offset: 41, CombiningMark: ""},
	'q': {Offset: 42, CombiningMark: ""},
	'r': {Offset: 43, CombiningMark: ""},
	's': {Offset: 44, CombiningMark: ""},
	't': {Offset: 45, CombiningMark: ""},
	'u': {Offset: 46, CombiningMark: ""},
	'v': {Offset: 47, CombiningMark: ""},
	'w': {Offset: 48, CombiningMark: ""},
	'x': {Offset: 49, CombiningMark: ""},
	'y': {Offset: 50, CombiningMark: ""},
	'z': {Offset: 51, CombiningMark: ""},
	'À': {Offset: 0, CombiningMark: "\u0300"},
	'à': {Offset: 26, CombiningMark: "\u0300"},
	'Á': {Offset: 0, CombiningMark: "\u0301"},
	'á': {Offset: 26, CombiningMark: "\u0301"},
	'Â': {Offset: 0, CombiningMark: "\u0302"},
	'â': {Offset: 26, CombiningMark: "\u0302"},
	'Ã': {Offset: 0, CombiningMark: "\u0303"},
	'ã': {Offset: 26, CombiningMark: "\u0303"},
	'Å': {Offset: 0, CombiningMark: "\u030A"},
	'å': {Offset: 26, CombiningMark: "\u030A"},
	'Ç': {Offset: 2, CombiningMark: "\u0327"},
	'ç': {Offset: 28, CombiningMark: "\u0327"},
	'É': {Offset: 4, CombiningMark: "\u0301"},
	'é': {Offset: 30, CombiningMark: "\u0301"},
	'Ê': {Offset: 4, CombiningMark: "\u0302"},
	'ê': {Offset: 30, CombiningMark: "\u0302"},
	'Ë': {Offset: 4, CombiningMark: "\u0308"},
	'ë': {Offset: 30, CombiningMark: "\u0308"},
	'Í': {Offset: 8, CombiningMark: "\u0301"},
	'í': {Offset: 34, CombiningMark: "\u0301"},
	'Î': {Offset: 8, CombiningMark: "\u0302"},
	'î': {Offset: 34, CombiningMark: "\u0302"},
	'Ï': {Offset: 8, CombiningMark: "\u0308"},
	'ï': {Offset: 34, CombiningMark: "\u0308"},
	'Ñ': {Offset: 13, CombiningMark: "\u0303"},
	'ñ': {Offset: 39, CombiningMark: "\u0303"},
	'Ó': {Offset: 14, CombiningMark: "\u0301"},
	'ó': {Offset: 40, CombiningMark: "\u0301"},
	'Ô': {Offset: 14, CombiningMark: "\u0302"},
	'ô': {Offset: 40, CombiningMark: "\u0302"},
	'Ö': {Offset: 14, CombiningMark: "\u0308"},
	'ö': {Offset: 40, CombiningMark: "\u0308"},
	'Ù': {Offset: 20, CombiningMark: "\u0300"},
	'ù': {Offset: 46, CombiningMark: "\u0300"},
	'Ú': {Offset: 20, CombiningMark: "\u0301"},
	'ú': {Offset: 46, CombiningMark: "\u0301"},
	'Û': {Offset: 20, CombiningMark: "\u0302"},
	'û': {Offset: 46, CombiningMark: "\u0302"},
	'Ü': {Offset: 20, CombiningMark: "\u0308"},
	'ü': {Offset: 46, CombiningMark: "\u0308"},
	'Ø': {Offset: 14, CombiningMark: "\u0338"},
	'ø': {Offset: 40, CombiningMark: "\u0338"},
}

var numbers = map[rune]string{
	'0': "\U0001D7F6", '1': "\U0001D7F7", '2': "\U0001D7F8", '3': "\U0001D7F9", '4': "\U0001D7FA",
	'5': "\U0001D7FB", '6': "\U0001D7FC", '7': "\U0001D7FD", '8': "\U0001D7FE", '9': "\U0001D7FF",
}

var boldNumbers = map[rune]string{
	'0': "\U0001D7CE", '1': "\U0001D7CF", '2': "\U0001D7D0", '3': "\U0001D7D1", '4': "\U0001D7D2",
	'5': "\U0001D7D3", '6': "\U0001D7D4", '7': "\U0001D7D5", '8': "\U0001D7D6", '9': "\U0001D7D7",
}

func Format(input string) string {
	return FormatWithStyle(input, Regular)
}

func FormatWithStyle(input string, style FontStyle) string {
	numbersMap := numbers
	if style == Bold || style == BoldItalic {
		numbersMap = boldNumbers
	}
	var result strings.Builder
	for _, char := range input {
		if mapping, ok := letterMap[char]; ok {
			result.WriteString(string(rune(int32(style) + mapping.Offset)))
			if mapping.CombiningMark != "" {
				result.WriteString(mapping.CombiningMark)
			}
		} else if replacement, ok := numbersMap[char]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}
