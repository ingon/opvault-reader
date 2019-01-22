package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	// "github.com/Ingon/1passread/opvault"
	"github.com/atotto/clipboard"
	"github.com/miquella/opvault"
)

type Secret struct {
	name  string
	value string
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	location := os.Args[1]
	pass := os.Args[2]
	regex := regexp.MustCompile(fmt.Sprintf("(?i).*%s.*", os.Args[3]))

	vault, err := opvault.Open(location)
	check(err)

	profile, err := vault.Profile("default")
	check(err)

	err = profile.Unlock(pass)
	check(err)

	items, err := profile.Items()
	check(err)

	var matchedItems []*opvault.Item
	for _, item := range items {
		if regex.MatchString(item.Title()) {
			matchedItems = append(matchedItems, item)
		}
	}

	item := chooseItem(matchedItems)

	detail, err := item.Detail()
	check(err)

	var secrets []Secret

	for _, field := range detail.Fields() {
		val := field.Value()

		if field.Type() == opvault.PasswordFieldType {
			secrets = append(secrets, Secret{
				name:  field.Name(),
				value: val,
			})
			val = "*******************"
		}

		fmt.Printf(" - %s (%s)/%s: %s\n", field.Name(), field.Type(), field.Designation(), val)
	}

	for _, section := range detail.Sections() {
		fmt.Printf(" + %s (%s)\n", section.Title(), section.Name())
		for _, field := range section.Fields() {
			val := field.Value()

			if field.Kind() == opvault.ConcealedFieldKind {
				secrets = append(secrets, Secret{
					name:  field.Name(),
					value: val,
				})
				val = "*******************"
			}

			fmt.Printf(" -- %s (%s)/%s: %s\n", field.Title(), field.Name(), field.Kind(), val)
		}
	}

	secret, err := chooseSecret(secrets)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	current, err := clipboard.ReadAll()
	if err != nil {
		current = ""
	}

	err = clipboard.WriteAll(secret.value)
	check(err)

	ch := make(chan struct{})
	go func() {
		fmt.Scanln()
		close(ch)
	}()

	select {
	case <-ch:
		break
	case <-time.After(10 * time.Second):
		break
	}

	err = clipboard.WriteAll(current)
	check(err)
}

func chooseItem(matchedItems []*opvault.Item) *opvault.Item {
	switch len(matchedItems) {
	case 0:
		panic("No items")
	case 1:
		return matchedItems[0]
	}

	for k, item := range matchedItems {
		fmt.Printf("%d) %s\n", k+1, item.Title())
	}

	var num int
	fmt.Printf("> ")
	_, err := fmt.Scanf("%d", &num)
	check(err)

	return matchedItems[num-1]
}

func chooseSecret(secrets []Secret) (Secret, error) {
	switch len(secrets) {
	case 0:
		return Secret{}, fmt.Errorf("No secrets")
	case 1:
		return secrets[0], nil
	}

	for k, sec := range secrets {
		fmt.Printf("%d) %s\n", k+1, sec.name)
	}

	var num int
	fmt.Printf("> ")
	_, err := fmt.Scanf("%d", &num)
	check(err)

	return secrets[num-1], nil
}
