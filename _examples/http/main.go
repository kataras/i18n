package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/kataras/i18n"
)

type user struct {
	Name string
	Age  int
}

func main() {
	// Custom I18n instance.
	I18n, err := i18n.New(i18n.Glob("../basic/locales/*/*"), "en-US", "el-GR", "zh-CN")
	if err != nil {
		panic(err)
	}

	I18n.URLParameter = "lang" // e.g. ?lang=el
	I18n.Cookie = "lang"

	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		text := I18n.GetMessage(r, "hi", "John Doe")
		fmt.Fprintf(w, "Text: %s", text)
	})

	router.HandleFunc("/some-path", func(w http.ResponseWriter, r *http.Request) {
		templateData := user{
			Name: "John Doe",
			Age:  66,
		}

		text := I18n.GetMessage(r, "intro", templateData)
		fmt.Fprintf(w, "Text: %s", text)
	})

	htmlTmpl, err := template.New("index.html").Parse(`Text: <strong> {{call .tr "hi" "John Doe"}} </strong>`)
	if err != nil {
		panic(err)
	}

	router.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		htmlTmpl.Execute(w, map[string]interface{}{
			// format, args, see: `{{ call .tr "key/format" arguments... }}`
			"tr": I18n.GetLocale(r).GetMessage,
		})
	})

	// Optionally, wrap the router with `I18n.Router` to set the language
	// based on a subdomain or path prefix WITHOUT add new routes for each language.
	//
	// go to http://localhost:8080/el-gr/some-path (by path prefix)
	// or http://el.mydomain.com8080/some-path (by subdomain - test locally with the hosts file)
	// or http://localhost:8080/zh-CN/templates (by path prefix with uppercase)
	// or http://localhost:8080/some-path?lang=el-GR (by url parameter)
	// or http://localhost:8080 (default is en-US)
	// or http://localhost:8080/?lang=zh-CN
	//
	// or use cookies to set the language.
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", I18n.Router(router))
}
