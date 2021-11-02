package main

import (
	"fmt"
)

func getExternals() []string {
	defaultExternals := []string{"path", "fs", "crypto", "os", "http", "child_process", "querystring", "readline", "tls", "assert", "buffer", "url", "net", "buffer", "tty", "util", "stream", "events", "zlib", "https", "worker_threads", "module", "http2", "dns"}

	cliExternals := externals.Value()

	if len(cliExternals) > 0 {
		fmt.Println("Cli externals", externals.Value())
		defaultExternals = append(defaultExternals, cliExternals...)
	}

	return defaultExternals
}
