package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kataras/i18n"
)

const (
	female = iota + 1
	male
)

const tableStyle = `
<style>
a {
    padding: 8px 8px;
    text-decoration:none;
    cursor:pointer;
    color: #10a2ff;
}
table {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    height: 100%;
    width: 100%;
    border-collapse: collapse;
    border-spacing: 0;
    empty-cells: show;
    border: 1px solid #cbcbcb;
}

table caption {
    color: #000;
    font: italic 85%/1 arial, sans-serif;
    padding: 1em 0;
    text-align: center;
}

table td,
table th {
    border-left: 1px solid #cbcbcb;
    border-width: 0 0 0 1px;
    font-size: inherit;
    margin: 0;
    overflow: visible;
    padding: 0.5em 1em;
}

table thead {
    background-color: #10a2ff;
    color: #fff;
    text-align: left;
    vertical-align: bottom;
}

table td {
    background-color: transparent;
}

.table-odd td {
    background-color: #f2f2f2;
}

.table-bordered td {
    border-bottom: 1px solid #cbcbcb;
}
.table-bordered tbody > tr:last-child > td {
    border-bottom-width: 0;
}
</style>
`

// I18n our i18n instance. We set it as package-level variable for the sake of the example.
// Note that, we use 1 language in this example but you can extend to as many as you want.
var I18n, _ = i18n.New(i18n.Glob("./locales/*/*"), "en-US")

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "<html><body>\n")
		io.WriteString(w, tableStyle)
		io.WriteString(w, `<table class="table-bordered table-odd">
<thead>
  <tr>
    <th>Key</th>
    <th>Translation</th>
    <th>Arguments</th>
  </tr>
</thead><tbody>
`)
		defer io.WriteString(w, "</tbody></table></body></html>")

		tr(w, r, "Classic")

		tr(w, r, "YouLate", 1)
		tr(w, r, "YouLate", 2)

		tr(w, r, "FreeDay", 1)
		tr(w, r, "FreeDay", 5)

		tr(w, r, "FreeDay", 3, 15)

		tr(w, r, "HeIsHome", "Peter")

		tr(w, r, "HouseCount", female, 2, "Maria")
		tr(w, r, "HouseCount", male, 1, "Peter")

		tr(w, r, "nav.home")
		tr(w, r, "nav.user")
		tr(w, r, "nav.more.what")
		tr(w, r, "nav.more.even.more")
		tr(w, r, "nav.more.even.aplural", 1)
		tr(w, r, "nav.more.even.aplural", 15)

		tr(w, r, "VarTemplate", map[string]interface{}{
			"Name":        "Peter",
			"GenderCount": male,
		})

		tr(w, r, "VarTemplatePlural", 1, female)
		tr(w, r, "VarTemplatePlural", 2, female, 1)
		tr(w, r, "VarTemplatePlural", 2, female, 5)
		tr(w, r, "VarTemplatePlural", 1, male)
		tr(w, r, "VarTemplatePlural", 2, male, 1)
		tr(w, r, "VarTemplatePlural", 2, male, 2)

		tr(w, r, "VarTemplatePlural", map[string]interface{}{
			"PluralCount": 5,
			"Names":       []string{"Makis", "Peter"},
			"InlineJoin": func(arr []string) string {
				return strings.Join(arr, ", ")
			},
		})

		tr(w, r, "TemplatePlural", map[string]interface{}{
			"PluralCount": 1,
			"Name":        "Peter",
		})
		tr(w, r, "TemplatePlural", map[string]interface{}{
			"PluralCount": 5,
			"Names":       []string{"Makis", "Peter"},
			"InlineJoin": func(arr []string) string {
				return strings.Join(arr, ", ")
			},
		})
		tr(w, r, "VarTemplatePlural", 2, male, 4)

		tr(w, r, "TemplateVarTemplatePlural", map[string]interface{}{
			"PluralCount": 3,
			"DogsCount":   5,
		})

		tr(w, r, "message.HostResult")

		tr(w, r, "LocalVarsHouseCount.Text", 3, 4)
	})

	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}

func tr(w http.ResponseWriter, r *http.Request, key string, args ...interface{}) {
	translation := I18n.GetLocale(r).GetMessage(key, args...)
	fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%v</td></tr>\n", key, translation, args)
}
