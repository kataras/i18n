package main

import (
	"fmt"
	"text/template"

	"github.com/kataras/i18n"

	// go get -u github.com/gertd/go-pluralize
	"github.com/gertd/go-pluralize"
)

/*
I18n supports text/template inside the translation values.
Follow this example to learn how to use that feature.

This is just an example on how to use template functions.
See the "plurals" example for a more comprehensive pluralization support instead.
*/
var pluralizeClient = pluralize.NewClient()

func getFuncs(current *i18n.Locale) template.FuncMap {

	return template.FuncMap{
		"plural": func(word string, count int) string {
			// Your own implementation or use a 3rd-party package
			// like we do here.
			//
			// Note that this is only for english,
			// but you can use the "current" locale
			// and make a map with dictionaries to
			// pluralize words based on the given language.
			return pluralizeClient.Pluralize(word, count, true)
		},
	}
}

func main() {
	I18n, err := i18n.New(i18n.Glob("./locales/*/*", i18n.LoaderConfig{
		// Set custom functions per locale!
		Funcs: getFuncs,
	}), "en-US", "el-GR", "zh-CN")
	if err != nil {
		panic(err)
	}

	textEnglish := I18n.Tr("en", "HiDogs", map[string]interface{}{
		"count": 2,
	}) // prints "Hi 2 dogs".
	fmt.Println(textEnglish)

	textEnglishSingular := I18n.Tr("en", "HiDogs", map[string]interface{}{
		"count": 1,
	}) // prints "Hi 1 dog".
	fmt.Println(textEnglishSingular)

	textGreek := I18n.Tr("el", "HiDogs", map[string]interface{}{
		"count": 1,
	}) // prints "Γειά 1 σκυλί".
	fmt.Println(textGreek)
}
