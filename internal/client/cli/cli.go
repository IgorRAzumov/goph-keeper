package cli

import (
	"fmt"
	"io"
	"os"
)

// Run выполняет команду CLI (os.Args[1:]).
func Run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return 2
	}
	switch args[0] {
	case "version":
		return runVersion()
	case "register":
		return runRegister(args[1:])
	case "login":
		return runLogin(args[1:])
	case "logout":
		return runLogout()
	case "sync":
		return runSync()
	case "add":
		return runAdd(args[1:])
	case "list":
		return runList(args[1:])
	case "get":
		return runGet(args[1:])
	case "delete":
		return runDelete(args[1:])
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage(os.Stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `GophKeeper CLI

Usage:
  gophkeeper version
  gophkeeper register -login USER -password PASS [-server URL]
  gophkeeper login -login USER [-password PASS] [-server URL]
  gophkeeper logout
  gophkeeper sync
  gophkeeper add -type text|login|card|binary -meta META [options]
  gophkeeper list [-sync]
  gophkeeper get [-sync] ID
  gophkeeper delete ID

Environment:
  GOPHKEEPER_SERVER_URL       API base URL (default http://127.0.0.1:8080)
  GOPHKEEPER_MASTER_PASSWORD  master password for E2E encryption
  GOPHKEEPER_CONFIG_DIR       config directory (default ~/.gophkeeper)

`)
}
