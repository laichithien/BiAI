//go:build !windows

package platform

import "log"

func ShowError(title, message string) {
	log.Printf("%s: %s", title, message)
}
