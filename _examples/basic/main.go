package main

import (
	"fmt"

	"github.com/kataras/i18n"
)

type user struct {
	Name string
	Age  int
}

func main() {
	// i18n.SetDefaultLanguage("en-US")

	// Fmt style.
	enText := i18n.Tr("en", "hi", "John Doe") // or "en-US"
	elText := i18n.Tr("el", "hi", "John Doe")
	zhText := i18n.Tr("zh", "hi", "John Doe")

	fmt.Println(enText)
	fmt.Println(elText)
	fmt.Println(zhText)

	// Templates style.
	templateData := user{
		Name: "John Doe",
		Age:  66,
	}

	enText = i18n.Tr("en-US", "intro", templateData) // or "en"
	elText = i18n.Tr("el-GR", "intro", templateData)
	zhText = i18n.Tr("zh-CN", "intro", templateData)

	fmt.Println(enText)
	fmt.Println(elText)
	fmt.Println(zhText)
}
