# i18n (Go)

[![build status](https://img.shields.io/travis/kataras/i18n/master.svg?style=for-the-badge&logo=travis)](https://travis-ci.org/kataras/i18n) [![report card](https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=for-the-badge)](https://goreportcard.com/report/github.com/kataras/i18n) [![godocs](https://img.shields.io/badge/go-%20docs-488AC7.svg?style=for-the-badge)](https://godoc.org/github.com/kataras/i18n) [![donate on PayPal](https://img.shields.io/badge/support-PayPal-blue.svg?style=for-the-badge)](https://www.paypal.me/kataras)

Efficient and easy to use localization and internationalization support for Go.

## Getting started

The only requirement is the [Go Programming Language](https://golang.org/dl).

```sh
$ go get github.com/kataras/i18n
```

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

- _en.file.yaml_
- _file_en-US.json_
- _/el/file.tml_

The [Default](https://github.com/kataras/i18n/blob/master/i18n.go#L37) `I18n` instance will try to load locale files from `./locales` directory.
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

Set the translate function as a key on a `Template`.

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
