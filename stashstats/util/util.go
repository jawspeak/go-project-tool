package util

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

func FatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// returns true if there was an error
func FatalIfErrUnless(err error, nonFatalIf func(err error) bool, usedParams interface{}) bool {
	if err != nil {
		fmt.Println("error: ", spew.Sdump(err), "attempted with params: ", spew.Sdump(usedParams))
		if !nonFatalIf(err) {
			log.Fatal(err)
		}
		return true
	}
	return false
}
