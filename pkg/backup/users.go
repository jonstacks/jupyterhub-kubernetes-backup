package backup

import (
	"fmt"
	"regexp"
	"strings"
)

var userDomainRegex = regexp.MustCompile(`(.*?)-40\w+-com`)

// GetUserNameFromPVCName returns the username formatted as user.name or user-name
func GetUserNameFromPVCName(s string) string {
	s = strings.TrimPrefix(s, "claim-")
	match := userDomainRegex.FindStringSubmatch(s)
	if len(match) > 1 {
		fmt.Printf("%v", match)
		s = match[1]
	}
	for _, encode := range []string{"-20", "-2e", "-40"} {
		if strings.Contains(s, encode) {
			s = strings.ReplaceAll(s, encode, ".")
		}
	}
	return s
}
