package cli

import (
	"context"
	"fmt"
	"os"

	clientapp "goph-keeper/internal/client/app"
)

func runRegister(args []string) int {
	flagSet := newFlagSet("register")
	login := flagSet.String("login", "", "account login")
	password := flagSet.String("password", "", "account password")
	server := flagSet.String("server", "", "server URL")
	if flagSet.Parse(args) != nil {
		return 2
	}
	if *login == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "register: -login and -password required")
		return 2
	}
	master, err := readMasterPassword()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	app, err := clientapp.New(master)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := app.Register(context.Background(), *login, *password, *server); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("registered and logged in")
	return 0
}

func runLogin(args []string) int {
	flagSet := newFlagSet("login")
	login := flagSet.String("login", "", "account login")
	password := flagSet.String("password", "", "account password")
	server := flagSet.String("server", "", "server URL")
	if flagSet.Parse(args) != nil {
		return 2
	}
	if *login == "" {
		fmt.Fprintln(os.Stderr, "login: -login required")
		return 2
	}
	pass := *password
	if pass == "" {
		var err error
		pass, err = readPassword("account password: ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	master, err := readMasterPassword()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	app, err := clientapp.New(master)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := app.Login(context.Background(), *login, pass, *server); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("logged in")
	return 0
}

func runLogout() int {
	app, err := clientapp.New("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := app.Logout(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("logged out")
	return 0
}
