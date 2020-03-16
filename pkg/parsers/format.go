package parsers

import "regexp"

var (
	quakeColorRE = regexp.MustCompile(`(\^\d)+`)
)

// RemoveQ3Colors removes all special color formatting returned from the Quake 3 server.
func RemoveQ3Colors(str string) string {
	return quakeColorRE.ReplaceAllString(str, "")
}
