package util

import "log"

func FatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// returns true if there was an error
func FatalIfErrUnless(err error, nonFatalIf func(err error) bool) bool {
	if err != nil {
		if !nonFatalIf(err) {
			log.Fatal(err)
		}
		return true
	}
	return false
}
