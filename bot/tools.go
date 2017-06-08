package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//QuotedArrayToString takes an array, and extracts the quoted string.
func QuotedArrayToString(array []string) (quote string) {
	var inner bool
	for _, p := range array {
		if strings.HasPrefix(p, "\"") && inner == false {
			quote += p + " "
			inner = true
		} else if inner {
			quote += p + " "
		} else if strings.HasSuffix(p, "\"") && inner {
			quote += p
			inner = false
		}
	}

	if inner {
		// Potentially handle bad quoted.
	}

	return
}

func tsConvert(ts discordgo.Timestamp) string {
	a := strings.FieldsFunc(fmt.Sprintf("%s", ts), tsSplit)
	return fmt.Sprintf("%s %s", a[0], a[1])
}

func tsSplit(r rune) bool {
	return r == 'T' || r == '.' || r == '+'
}
