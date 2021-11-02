package main

import (
	"sync"
)

var modules = make(chan string, 100)

var moduleWg sync.WaitGroup

var count int32

func checkModule(name string) {
	moduleWg.Add(1)
	count += 1
	// fmt.Println("Added", name, "to check queue")
	modules <- name
}

func handleModule() {
	for name := range modules {
		if _, ok := deps[name]; ok {
			deps[name] = true
		}
		moduleWg.Done()
		count -= 1
		// fmt.Println("Removed", name, "from check queue. There are ", count, "items remaining")
	}

}
