package utils

import (
	"regexp"
	"strings"
)

func SanitizeUsername(username string) (string, bool) {
	u := strings.ToLower(strings.TrimSpace(username))

	u = strings.ReplaceAll(u, " ", "")

	re := regexp.MustCompile(`^[a-z][a-z0-9_]{2,19}$`)

	if !re.MatchString(u) {
		return "", false
	}

	return u, true
}

func SanitizePhone(phone string) (string, bool) {
	p := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "").Replace(phone)

	if strings.HasPrefix(p, "00225") {
		p = strings.TrimPrefix(p, "00225")
	} else if strings.HasPrefix(p, "225") {
		p = strings.TrimPrefix(p, "225")
	}

	if len(p) != 10 {
		return "", false
	}

	if !strings.HasPrefix(p, "01") && !strings.HasPrefix(p, "05") && !strings.HasPrefix(p, "07") {
		return "", false
	}

	re := regexp.MustCompile(`^[0-9]+$`)
	if !re.MatchString(p) {
		return "", false
	}

	return p, true
}
