package main

import (
	"runtime"
)

func main() {
	run()
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
