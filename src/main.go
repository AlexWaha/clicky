package main

import "runtime"

func main() {
	runtime.LockOSThread()
	initApp()
	platformRun()
}
