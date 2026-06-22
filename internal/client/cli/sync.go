package cli

import (
	"context"
	"fmt"
	"os"
)

func runSync() int {
	app, err := openApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := app.Sync(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("sync ok")
	return 0
}
