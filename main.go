/*
Specdeck manages a local directory of specifications and builds complete
export artifacts from them.

Usage:

	specdeck <command> [flags]

Commands:

	new        Initialise a new specdeck project in the current directory
	link       Link a code repository to a specdeck specs repo
	agents     Install or update specdeck's commands for AI coding agents
	build      Resolve container refs and write flat specs to dist/
*/
package main

import "github.com/hdirksor/specdeck/cmd"

func main() {
	cmd.Execute()
}
