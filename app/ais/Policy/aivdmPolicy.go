package Policy

import "fmt"

func isMessageAivdm(prefix string) bool {
	if prefix != "!AIVDM" {
		return false
	}

	return true
}

func AssertMessageIsAivdm(prefix string) error {
	if !isMessageAivdm(prefix) {
		return fmt.Errorf("invalid prefix: %s", prefix)
	}

	return nil
}
