package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kataras/i18n"
)

const (
	female = iota + 1
	male
)

var I18n, _ = i18n.New(i18n.Glob("./locales/*/*"), "en-US")

func TestI18nPlurals(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
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

		tr(w, r, "VarTemplate", i18n.Map{
			"Name":        "Peter",
			"GenderCount": male,
		})

		tr(w, r, "VarTemplatePlural", 1, female)
		tr(w, r, "VarTemplatePlural", 2, female, 1)
		tr(w, r, "VarTemplatePlural", 2, female, 5)
		tr(w, r, "VarTemplatePlural", 1, male)
		tr(w, r, "VarTemplatePlural", 2, male, 1)
		tr(w, r, "VarTemplatePlural", 2, male, 2)

		tr(w, r, "VarTemplatePlural", i18n.Map{
			"PluralCount": 5,
			"Names":       []string{"Makis", "Peter"},
			"InlineJoin": func(arr []string) string {
				return strings.Join(arr, ", ")
			},
		})

		tr(w, r, "TemplatePlural", i18n.Map{
			"PluralCount": 1,
			"Name":        "Peter",
		})
		tr(w, r, "TemplatePlural", i18n.Map{
			"PluralCount": 5,
			"Names":       []string{"Makis", "Peter"},
			"InlineJoin": func(arr []string) string {
				return strings.Join(arr, ", ")
			},
		})
		tr(w, r, "VarTemplatePlural", 2, male, 4)

		tr(w, r, "TemplateVarTemplatePlural", i18n.Map{
			"PluralCount": 3,
			"DogsCount":   5,
		})

		tr(w, r, "message.HostResult")

		tr(w, r, "LocalVarsHouseCount.Text", 3, 4)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	defer r.Body.Close()
	handler(w, r)

	expected := `Classic=classic
YouLate=You are 1 minute late.
YouLate=You are 2 minutes late.
FreeDay=You have a day off
FreeDay=You have 5 free days
FreeDay=You have three days and 15 minutes off.
HeIsHome=Peter is home
HouseCount=She (Maria) has 2 houses
HouseCount=He (Peter) has 1 house
nav.home=Home
nav.user=Account
nav.more.what=this
nav.more.even.more=yes
nav.more.even.aplural=You are 1 minute late.
nav.more.even.aplural=You are 15 minutes late.
VarTemplate=(He) Peter is home
VarTemplatePlural=She is awesome
VarTemplatePlural=other (She) has 1 house
VarTemplatePlural=other (She) has 5 houses
VarTemplatePlural=He is awesome
VarTemplatePlural=other (He) has 1 house
VarTemplatePlural=other (He) has 2 houses
VarTemplatePlural=Makis, Peter are awesome
TemplatePlural=Peter is unique
TemplatePlural=Makis, Peter are awesome
VarTemplatePlural=other (He) has 4 houses
TemplateVarTemplatePlural=These 3 are wonderful, feeding 5 dogsssss in total!
message.HostResult=Store Encrypted Message Online
LocalVarsHouseCount.Text=She has 4 houses
`
	if got := w.Body.String(); expected != got {
		t.Fatalf("expected:\n'%s'\n\nbut got:\n'%s'", expected, got)
	}
}

func tr(w http.ResponseWriter, r *http.Request, key string, args ...interface{}) {
	translation := I18n.GetLocale(r).GetMessage(key, args...)
	fmt.Fprintf(w, "%s=%s\n", key, translation)
}
