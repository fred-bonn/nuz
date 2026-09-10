package engine

import (
	"fmt"
	"log"
)

// Verbose controls whether vprint* helpers emit output; set by callers of the engine package.
var Verbose bool

func vprintln(v ...any) {
	if !Verbose {
		return
	}
	fmt.Println(v...)
}

func vprintf(format string, v ...any) {
	if !Verbose {
		return
	}
	fmt.Printf(format+"\n", v...)
}

func vprintWithPrefix(prefix, format string, v ...any) {
	if !Verbose {
		return
	}
	fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, v...))
}

func vprintMove(prio, speed int, format string, v ...any) {
	vprintWithPrefix(fmt.Sprintf("[MOVE, %d/%d]", prio, speed), format, v...)
}

func vprintSwitch(format string, v ...any) {
	vprintWithPrefix("[SWITCH]", format, v...)
}

func vprintReplace(format string, v ...any) {
	vprintWithPrefix("[REPLACE]", format, v...)
}

func vprintItem(format string, v ...any) {
	vprintWithPrefix("[ITEM]", format, v...)
}

func elogf(format string, v ...any) {
	log.Printf(format, v...)
}
