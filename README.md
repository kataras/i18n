# i18n (Go)

[![build status](https://img.shields.io/travis/kataras/i18n/master.svg?style=for-the-badge&logo=travis)](https://travis-ci.org/kataras/i18n) [![report card](https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=for-the-badge)](https://goreportcard.com/report/github.com/kataras/i18n) [![godocs](https://img.shields.io/badge/go-%20docs-488AC7.svg?style=for-the-badge)](https://godoc.org/github.com/kataras/i18n) [![donate on Stripe](https://img.shields.io/badge/support-Stripe-blue.svg?style=for-the-badge)](https://iris-go.com/donate)

Efficient and easy to use localization and internationalization support for Go.

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl).

```sh
$ go get -u github.com/kataras/i18n
```

**Examples**

- [Basic](_examples/basic)
- [Template](_examples/template)
- [Pluralization](_examples/plurals) **NEW**
    - [en-US/welcome.yml](_examples/plurals/locales/en-US/welcome.yml)
    - [en-US/ini_example.ini](_examples/plurals/locales/en-US/ini_example.ini)
- [HTTP](_examples/http)
- [Embedded Locales](_examples/embedded-files)

## Getting started

Create a folder named `./locales` and put some `YAML`, `TOML`, `JSON` or `INI` files.

```sh
│   main.go
└───locales
    ├───el-GR
    │       example.yml
    ├───en-US
    │       example.yml
    └───zh-CN
            example.yml
```

Now, put the key-values content for each locale, e.g. **./locales/en-US/example.yml** 

```yaml
hi: "Hi %s"
#
# Templates are supported
# hi: "Hi {{ .Name }}
#
# Template functions are supported
# hi: "Hi {{sayHi .Name}}
```

```yaml
# ./locales/el-GR/example.yaml
hi: "Γειά σου %s"
```

```yaml
# ./locales/zh-CN/example.yaml
hi: 您好 %s
```

Some other possible filename formats...

- _page.en.yaml_
- _home.cart.el-GR.json_
- _/el/file.tml_

> The language code MUST be right before the file extension.

The [Default](https://github.com/kataras/i18n/blob/master/i18n.go#L33) `I18n` instance will try to load locale files from `./locales` directory.
Use the `Tr` package-level function to translate a text based on the given language code. Use the `GetMessage` function to translate a text based on the incoming `http.Request`. Use the `Router` function to wrap an `http.Handler` (i.e an `http.ServeMux`) to set the language based on _path prefix_ such as `/zh-CN/some-path` and subdomains such as `zh.domain.com` **without the requirement of different routes per language**.

Let's take a look at the simplest usage of this package.

```go
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
```

Load specific languages over a **new I18n instance**. The default language is the first registered, in that case is the "en-US".

```go
I18n, err := i18n.New(i18n.Glob("./locales/*/*"), "en-US", "el-GR", "zh-CN")

// load embedded files through a go-bindata package
I18n, err := i18n.New(i18n.Assets(AssetNames, Asset), "en-US", "el-GR", "zh-CN")
```

## Template variables & functions

Using **template variables & functions** as values in your locale value entry via `LoaderConfig`.

We are going to use a 3rd-party package for plural and singular words. Note that this is only for english dictionary, but you can use the `"current"` `Locale` and make a map with dictionaries to pluralize words based on the given language.

Before we get started, install the necessary packages:

```sh
$ go get -u github.com/kataras/i18n
$ go get -u github.com/gertd/go-pluralize
```

Let's create two simple translation files for our example. The `./locales/en-US/welcome.yml` and `./locales/el-GR/welcome.yml` respectfully:

```yml
Dog: "dog"
HiDogs: Hi {{plural (tr "Dog") .count }}
```

```yml
Dog: "σκυλί"
HiDogs: Γειά {{plural (tr "Dog") .count }}
```

> The `tr` template function is a builtin function registered per locale. It returns the key's translated value. E.g. on english file the `tr "Dog"` returns the `Dog:`'s value: `"dog"` and on greek file it returns `"σκυλί"`. This function helps importing a key to another key to complete a sentence.

Now, create a `main.go` file and store the following contents:

```go
package main

import (
    "fmt"
    "text/template"

    "github.com/kataras/i18n"
    "github.com/gertd/go-pluralize"
)

var pluralizeClient = pluralize.NewClient()

func getFuncs(current *i18n.Locale) template.FuncMap {
    return template.FuncMap{
        "plural": func(word string, count int) string {
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
```

Use `go run main.go` to run our small Go program. The output should look like this:

```sh
Hi 2 dogs
Hi 1 dog
Γειά 1 σκυλί
```

## HTTP

HTTP, automatically searches for url parameter, cookie, custom function and headers for the current user language.

```go
mux := http.NewServeMux()

I18n.URLParameter = "lang" // i.e https://domain.com?lang=el
I18n.Cookie = "lang"
I18n.ExtractFunc = func(r *http.Request) string { /* custom logic */ }

mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    translated := I18n.GetMessage(r, "hi", "John Doe")
    fmt.Fprintf(w, "Text: %s", translated)
})
```

Prefer `GetLocale` if more than one `GetMessage` call.

```go
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    locale := I18n.GetLocale(r)
    translated := locale.GetMessage("hi", "John Doe")
    // [...some locale.GetMessage calls]
})
```

Optionally, identify the current language by subdomain or path prefix, e.g.
en.domain.com and domain.com/en or domain.com/en-US and e.t.c.

```go
I18n.Subdomain = true

http.ListenAndServe(":8080", I18n.Router(mux))
```

If the `ContextKey` field is not empty then the `Router` will set the current language.

```go
I18n.ContextKey = "lang" 

mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    currentLang := r.Context().Value("lang").(string)
    fmt.Fprintf(w, "Language: %s", currentLang)
})
```

Set the translate function as a key on a `HTML Template`.

```go
templates, _ := template.ParseGlob("./templates/*.html")

mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // per-request.
    translateFunc := I18n.GetLocale(r).GetMessage

    templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
        "tr": translateFunc,
    })

    // {{ call .tr "hi" "John Doe" }}
})
```

Global function with the language as its first input argument.

```go
translateLangFunc := I18n.Tr
templates.Funcs(template.FuncMap{
    "tr": translateLangFunc,
})

// {{ tr "en" "hi" "John Doe" }}
```

For a more detailed technical documentation you can head over to our [godocs](https://godoc.org/github.com/kataras/i18n). And for executable code you can always visit the [_examples](_examples) repository's subdirectory.

## License

kataras/i18n is free and open-source software licensed under the [MIT License](https://tldrlegal.com/license/mit-license).
