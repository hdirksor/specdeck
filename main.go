/*
Specdeck manages a local directory of specifications and builds complete
export artifacts from them.

Usage:

	specdeck <command> [flags]

Commands:

	new        Initialise a new specdeck project in the current directory
	link       Link a code repository to a specdeck specs repo
	build      Resolve container refs and write flat specs to dist/
	validate   Validate cross-references in the specdeck project
	change     Manage change records
	sync       Update Claude Code skills to the current specdeck version
*/
package main

import "github.com/hdirksor/specdeck/cmd"

func main() {
	cmd.Execute()
}
