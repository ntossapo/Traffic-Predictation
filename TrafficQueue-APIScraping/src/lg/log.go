package lg

import "fmt"

const debug = false

func PrintLog(s string, ss string){
	if debug {
		fmt.Printf("[%s]\t:\t%s\n", s, ss)
	}
}

func Debug(s string){
	if debug {
		PrintLog("Debug", s)
	}
}
