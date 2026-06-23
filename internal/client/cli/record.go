package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
)

func runAdd(args []string) int {
	flagSet := newFlagSet("add")
	recordType := flagSet.String("type", "text", "record type: text, login, card, binary")
	meta := flagSet.String("meta", "", "metadata")
	data := flagSet.String("data", "", "text payload")
	userLogin := flagSet.String("user", "", "login for type=login")
	userPass := flagSet.String("pass", "", "password for type=login")
	cardNumber := flagSet.String("number", "", "card number")
	cardHolder := flagSet.String("holder", "", "card holder")
	cardExpiry := flagSet.String("expiry", "", "card expiry")
	cardCVV := flagSet.String("cvv", "", "card cvv")
	filePath := flagSet.String("file", "", "file path for type=binary")
	if flagSet.Parse(args) != nil {
		return 2
	}
	app, err := openApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	input := clientapp.AddInput{
		Type:     model.RecordType(*recordType),
		Meta:     *meta,
		FilePath: *filePath,
	}
	switch model.RecordType(*recordType) {
	case model.RecordTypeText:
		input.Payload = *data
	case model.RecordTypeLogin:
		input.Payload = map[string]string{"login": *userLogin, "password": *userPass}
	case model.RecordTypeCard:
		input.Payload = map[string]string{
			"number": *cardNumber, "holder": *cardHolder,
			"expiry": *cardExpiry, "cvv": *cardCVV,
		}
	case model.RecordTypeBinary:
		// file or empty
	default:
		fmt.Fprintln(os.Stderr, "unsupported type")
		return 2
	}
	id, err := app.Add(context.Background(), input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(id)
	return 0
}

// openAppMaybeSync открывает app и, если задан флаг -sync, синхронизируется
// перед чтением.
func openAppMaybeSync(flagSet *flag.FlagSet) (*clientapp.App, error) {
	app, err := openApp()
	doSync := *flagSet.Bool("sync", false, "sync with server")

	if err != nil {
		return nil, err
	}
	if doSync {
		if err := app.Sync(context.Background()); err != nil {
			return nil, err
		}
	}
	return app, nil
}

func runList(args []string) int {
	flagSet := newFlagSet("list")
	if flagSet.Parse(args) != nil {
		return 2
	}
	app, err := openAppMaybeSync(flagSet)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	records, err := app.List(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, record := range records {
		fmt.Printf("%s\t%s\t%s\t%s\n", record.ID, record.Type, record.Meta, truncate(record.Payload, 40))
	}
	return 0
}

func runGet(args []string) int {
	flagSet := newFlagSet("get")
	if flagSet.Parse(args) != nil {
		return 2
	}
	id := flagSet.Arg(0)
	if id == "" {
		fmt.Fprintln(os.Stderr, "get: id required")
		return 2
	}
	app, err := openAppMaybeSync(flagSet)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	result, err := app.Get(context.Background(), id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("id: %s\n", result.ID)
	fmt.Printf("type: %s\n", result.Type)
	fmt.Printf("meta: %s\n", result.Meta)
	fmt.Printf("data: %s\n", result.Payload)
	return 0
}

func runDelete(args []string) int {
	if len(args) == 0 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "delete: id required")
		return 2
	}
	app, err := openApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := app.Delete(context.Background(), args[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("deleted")
	return 0
}
