package i18n

import (
	"io"
	"os"
	"path/filepath"
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

func TestLoadKV(t *testing.T) {
	m := LangMap{
		"en-US": Map{
			"buy":               `buy %d`,
			"cart.checkout":     `checkout - {{.Param}}`,
			"cart.after.thanks": `thanks`,
			//
			"JSONTemplateExample":  `value of {{.Value}}`,
			"TypeOf":               `type of %T`,
			"KeyOnlyOnDefaultLang": `value`,
			//
			"title": `Title`,
			"hi":    `Hi {{.Name}}`,
			"int":   `1`,
			"hello": `Hello %s`,
			//
			"welcome": `welcome`,
		},
		"el-GR": Map{
			"buy":               `αγοράστε %d`,
			"cart.checkout":     `ολοκλήρωση παραγγελίας - {{.Param}}`,
			"cart.after.thanks": `ευχαριστούμε`,
			//
			"JSONTemplateExample": `τιμή του {{.Value}}`,
			"TypeOf":              `τύπος %T`,
			//
			"title": `Τίτλος`,
			"hi":    `Γειά σου {{.Name}}`,
			"int":   `1`,
			//
			"welcome": `καλώς ήρθατε`,
		},
	}
	i18N, err := New(KV(m))
	if err != nil {
		t.Fatal(err)
	}
	i18N.SetDefault("en-US")

	// Test welcome.
	got := i18N.Tr("el-GR", "welcome")
	if expected := "καλώς ήρθατε"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	// Test the rest.
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

	// test fallback default language's key.
	got = i18N.Tr("el-GR", "KeyOnlyOnDefaultLang")
	if expected := "value"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}
}

func TestLoadSingleFiles(t *testing.T) {
	i18N, err := New(Glob("./testfiles/*.yml"))
	if err != nil {
		t.Fatal(err)
	}

	i18N.SetDefault("en-US")

	got := i18N.Tr("el-GR", "welcome")
	if expected := "καλώς ήρθατε"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	got = i18N.Tr("en-US", "welcome")
	if expected := "welcome"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

	// test default en-US.
	got = i18N.Tr("ch-ZN", "welcome")
	if expected := "welcome"; got != expected {
		t.Fatalf("expected %s but got %s", expected, got)
	}

}
func TestLoadEmptyTags(t *testing.T) {
	i18N, err := New(Glob("./testfiles/*/*"))
	if err != nil {
		t.Fatal(err)
	}

	i18N.SetDefault("en-US")
	testLoadAndTrHelper(t, i18N)
}

// https://github.com/kataras/i18n/issues/1
func TestLoadAbsDirWithPotentialLangCode(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "/opt/etc/rcacs/locales")
	if err := createIfNotExists(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	if err := copyDir("./testfiles", dir); err != nil {
		t.Fatalf("failed to copy testfiles to the temp dir: %s: %v", dir, err)
	}

	i18N, err := New(Glob(dir + "/*/*"))
	if err != nil {
		t.Fatal(err)
	}

	i18N.SetDefault("en-US")
	testLoadAndTrHelper(t, i18N)
}

func copyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDir(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, perm); err != nil {
			return err
		}
	}

	return nil
}
