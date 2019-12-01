package i18n

import (
	"testing"
)

// go test -vet=off -v
func TestLoad(t *testing.T) {
	i18N, err := New(Glob("./testfiles/*/*"), "en-US", "el-GR")
	if err != nil {
		t.Fatal(err)
	}

	testLoadAndTrHelper(t, i18N)
}

func testLoadAndTrHelper(t *testing.T, i18N *I18n) {
	t.Helper()

	got := i18N.Tr("el", "hi", map[string]string{"Name": "kataras"})
	if expected := "Γειά σου kataras"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en-US", "hi", map[string]string{"Name": "kataras"})
	if expected := "Hi kataras"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en-US", "hello", "kataras")
	if expected := "Hello kataras"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("el-GR", "int")
	if expected := "1"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en", "JSONTemplateExample", map[string]string{"Value": "value"})
	if expected := "value of value"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("el-GR", "TypeOf", "a string")
	if expected := "τύπος string"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en", "buy", 2)
	if expected := "buy 2"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("el-GR", "buy", 2)
	if expected := "αγοράστε 2"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en", "cart.checkout", map[string]string{"Param": "all"})
	if expected := "checkout - all"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("el-GR", "cart.checkout", map[string]string{"Param": "όλα"})
	if expected := "ολοκλήρωση παραγγελίας - όλα"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}
}

func TestLoadEmptyTags(t *testing.T) {
	i18N, err := New(Glob("./testfiles/*/*"))
	if err != nil {
		t.Fatal(err)
	}

	testLoadAndTrHelper(t, i18N)
}
