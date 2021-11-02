package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

var logs = make(chan string, 100)

var logging bool

const DEPCHECK_DIR = ".depp"

var loggerWg sync.WaitGroup

var fileOps sync.WaitGroup

func fileLog(a ...interface{}) {
	str := fmt.Sprintln(a...)
	// fmt.Println("Adding to log queue")
	loggerWg.Add(1)
	logs <- str
}

func createDirectory() {
	if logging || report || esbuildWrite {
		if _, err := os.Stat(DEPCHECK_DIR); os.IsNotExist(err) {

			err := os.Mkdir(DEPCHECK_DIR, 0755)
			if err != nil {
				panic(err)
			}
		}
	}
}

func removeDirectory() {
	if _, err := os.Stat(DEPCHECK_DIR); !os.IsNotExist(err) {
		// path/to/whatever exists
		err := os.RemoveAll(DEPCHECK_DIR)
		if err != nil {
			panic(err)
		}
		fmt.Println("Cleaned!")
	} else {
		fmt.Println("Nothing to clean")
	}

}

func writeLogsToFile() {

	if logging {
		fmt.Println("Will be logging output to .depcheck.log")
		// open output file
		fo, err := os.Create(DEPCHECK_DIR + "/.depcheck.log")
		fileOps.Add(1)
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		datawriter := bufio.NewWriter(fo)

		for line := range logs {
			// fmt.Println("Writing", line, "to file")
			_, err := datawriter.WriteString(line + "\n")
			if err != nil {
				panic(err)
			}
			loggerWg.Done()
		}
		datawriter.Flush()
		fo.Close()
		fileOps.Done()
	} else {
		for line := range logs {
			fmt.Sprintln(line) // Just ignore this line
			loggerWg.Done()

		}
	}
}
