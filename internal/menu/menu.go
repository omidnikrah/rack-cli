package menu

import (
	"fmt"
	"strconv"
)

type MenuItems struct {
	Label    string
	Action   func()
	SubTitle string
	SubItems func() []MenuItems
}

type Screen interface {
	ClearScreen()
	ReadChoice() string
	WaitForEnter()
}

func RenderMenu(title string, items []MenuItems, screen Screen) {
	renderMenu(title, items, true, screen)
}

func renderMenu(title string, items []MenuItems, isRoot bool, screen Screen) {
	for {
		screen.ClearScreen()

		fmt.Println(title)

		for i, item := range items {
			fmt.Printf("%d. %s\n", i+1, item.Label)
		}

		if isRoot {
			fmt.Println("0. Exit")
		} else {
			fmt.Println("0. Back")
		}

		choice := screen.ReadChoice()
		idx, err := strconv.Atoi(choice)

		if err != nil || idx < 0 || idx > len(items) {
			fmt.Println("Invalid, try again")
			screen.WaitForEnter()
			continue
		}

		if idx == 0 {
			return
		}

		selected := items[idx-1]
		if selected.SubItems != nil {
			renderMenu(selected.SubTitle, selected.SubItems(), false, screen)
		} else {
			selected.Action()
			screen.WaitForEnter()
		}
	}
}
