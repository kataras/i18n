package main

import (
	"embed"
	"fmt"

	"github.com/kataras/i18n"
)

//go:embed static/locales/*
var staticFS embed.FS

type user struct {
	Name string
	Age  int
}

func main() {
	loaderOpts := i18n.LoaderConfig{
		Left:               "{{",
		Right:              "}}",
		Strict:             false,
		DefaultMessageFunc: nil,
		Funcs:              nil,
	}
	loader, err := i18n.FS(staticFS, "./static/locales/*/*.yml", loaderOpts)
	if err != nil {
		panic(err)
	}

	I18n, err := i18n.New(loader, "en-US", "el-GR", "zh-CN")
	if err != nil {
		panic(err)
	}

	//
	// The rest as expected...
	//

	// Fmt style.
	enText := I18n.Tr("en", "hi", "John Doe") // or "en-US"
	elText := I18n.Tr("el", "hi", "John Doe")
	zhText := I18n.Tr("zh", "hi", "John Doe")

	fmt.Println(enText)
	fmt.Println(elText)
	fmt.Println(zhText)

	// Templates style.
	templateData := user{
		Name: "John Doe",
		Age:  66,
	}

	enText = I18n.Tr("en-US", "intro", templateData) // or "en"
	elText = I18n.Tr("el-GR", "intro", templateData)
	zhText = I18n.Tr("zh-CN", "intro", templateData)

	fmt.Println(enText)
	fmt.Println(elText)
	fmt.Println(zhText)
}
