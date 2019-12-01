// Package i18n provides internalization and localization features.
package i18n

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

// Default keeps a package-level pre-loaded `I18n` instance.
// The default glob pattern is "./locales/*/*" which accepts folder
// structure as:
// - ./locales
//   - el-GR
//     - filename.yaml
//     - filename.toml
//     - filename.json
//   - en-US
//     - ...
//   - zh-CN
//     - ...
//   - ...
//
// The default language depends on the first lookup, please use the package-level `SetDefaultLanguage`
// to set a default language as you are not able to customize the language lists from here.
//
// See `New` package-level function to declare a fresh new, customized, `I18n` instance.
var Default *I18n

func init() {
	Default, _ = New(Glob("./locales/*/*"))
}

// SetDefaultLanguage changes the default language of the `Default` `I18n` instance.
func SetDefaultLanguage(langCode string) bool {
	return Default.SetDefault(langCode)
}

type (
	// Loader accepts a `Matcher` and should return a `Localizer`.
	// Functions that implement this type should load locale files.
	Loader func(m *Matcher) (Localizer, error)

	// Localizer is the interface which returned from a `Loader`.
	// Types that implement this interface should be able to retrieve a `Locale`
	// based on the language index.
	Localizer interface {
		// GetLocale should return a valid `Locale` based on the language index.
		// It will always match the Loader.Matcher.Languages[index].
		// It may return the default language if nothing else matches based on custom localizer's criteria.
		GetLocale(index int) Locale
	}

	// Locale is the interface which returns from a `Localizer.GetLocale` method.
	// It serves the translations based on "key" or format. See `GetMessage`.
	Locale interface {
		// Tag returns the full language Tag attached to this Locale,
		// it should be unique across different Locales.
		Tag() *language.Tag
		// Language should return the exact language code of this `Locale`
		// that the user provided on `New` function.
		//
		// Same as `Tag().String()` but it's static.
		Language() string
		// GetMessage should return translated text baesd on the given "key".
		GetMessage(key string, args ...interface{}) string
	}
)

// I18n is the structure which keeps the i18n configuration and implements Localization and internationalization features.
type I18n struct {
	localizer Localizer
	matcher   *Matcher

	loader Loader
	mu     sync.Mutex

	// If not nil, this request's context key can be used to identify the current language.
	// The found language(in this case, by path or subdomain) will be also filled with the current language on `Router` method.
	ContextKey interface{}
	// ExtractFunc is the type signature for declaring custom logic
	// to extract the language tag name.
	ExtractFunc func(*http.Request) string
	// If not empty, it is language identifier by url query.
	URLParameter string
	// If not empty, it is language identifier by cookie of this name.
	Cookie string
	// If true then a subdomain can be a language identifier too.
	Subdomain bool
}

// makeTags converts language codes to language Tags.
func makeTags(languages ...string) (tags []language.Tag) {
	for _, lang := range languages {
		tag, err := language.Parse(lang)
		if err == nil && tag != language.Und {
			tags = append(tags, tag)
		}
	}

	return
}

// New returns a new `I18n` instance.
// It contains a `Router` wrapper to (local) redirect subdomains and path prefixes too.
//
// The "languages" input parameter is optional and if not empty then only these languages
// will be used for translations and the rest (if any) will be skipped.
// the first parameter of "loader" which lookups for translations inside files.
func New(loader Loader, languages ...string) (*I18n, error) {
	tags := makeTags(languages...)

	i := &I18n{
		loader: loader,
		matcher: &Matcher{
			strict:    len(tags) > 0,
			Languages: tags,
			matcher:   language.NewMatcher(tags),
		},
	}

	if err := i.reload(); err != nil {
		return nil, err
	}

	return i, nil
}

// reload loads the language files from the provided Loader,
// the `New` package-level function preloads those files already.
func (i *I18n) reload() error { // May be an exported function, if requested.
	i.mu.Lock()
	defer i.mu.Unlock()

	localizer, err := i.loader(i.matcher)
	if err != nil {
		return err
	}

	i.localizer = localizer
	return nil
}

// SetDefault changes the default language.
// Please avoid using this method; the default behavior will accept
// the first language of the registered tags as the default one.
func (i *I18n) SetDefault(langCode string) bool {
	t, err := language.Parse(langCode)
	if err != nil {
		return false
	}

	if tag, index, conf := i.matcher.Match(t); conf > language.Low {
		if l, ok := i.localizer.(interface {
			SetDefault(int) bool
		}); ok {
			if l.SetDefault(index) {
				tags := i.matcher.Languages
				// set the order
				tags[index] = tags[0]
				tags[0] = tag

				i.matcher.Languages = tags
				i.matcher.matcher = language.NewMatcher(tags)
				return true
			}
		}
	}

	return false
}

// Matcher implements the languae.Matcher.
// It contains the original language Matcher and keeps an ordered
// list of the registered languages for further use (see `Loader` implementation).
type Matcher struct {
	strict    bool
	Languages []language.Tag
	matcher   language.Matcher
}

var _ language.Matcher = (*Matcher)(nil)

// Match returns the best match for any of the given tags, along with
// a unique index associated with the returned tag and a confidence
// score.
func (m *Matcher) Match(t ...language.Tag) (language.Tag, int, language.Confidence) {
	return m.matcher.Match(t...)
}

// MatchOrAdd acts like Match but it checks and adds a language tag, if not found,
// when the `Matcher.strict` field is true (when no tags are provided by the caller)
// and they should be dynamically added to the list.
func (m *Matcher) MatchOrAdd(t language.Tag) (tag language.Tag, index int, conf language.Confidence) {
	tag, index, conf = m.Match(t)
	if conf <= language.Low && !m.strict {
		// not found, add it now.
		m.Languages = append(m.Languages, t)
		tag = t
		index = len(m.Languages) - 1
		conf = language.Exact
		m.matcher = language.NewMatcher(m.Languages) // reset matcher to include the new language.
	}

	return
}

// ParseLanguageFiles returns a map of language indexes and
// their associated files based on the "fileNames".
func (m *Matcher) ParseLanguageFiles(fileNames []string) (map[int][]string, error) {
	languageFiles := make(map[int][]string)

	for _, fileName := range fileNames {
		index := parsePath(m, fileName)
		if index == -1 {
			continue
		}

		languageFiles[index] = append(languageFiles[index], fileName)
	}

	return languageFiles, nil
}

func parsePath(m *Matcher, path string) int {
	if t, ok := parseLanguage(path); ok {
		if _, index, conf := m.MatchOrAdd(t); conf > language.Low {
			return index
		}
	}

	return -1
}

func parseLanguage(path string) (language.Tag, bool) {
	if idx := strings.LastIndexByte(path, '.'); idx > 0 {
		path = path[0:idx]
	}

	// path = strings.ReplaceAll(path, "..", "")

	names := strings.FieldsFunc(path, func(r rune) bool {
		return r == '_' || r == os.PathSeparator || r == '/' || r == '.'
	})

	for _, s := range names {
		t, err := language.Parse(s)
		if err != nil {
			continue
		}

		return t, true
	}

	return language.Und, false
}

// TryMatchString will try to match the "s" with a registered language tag.
// It returns -1 as the language index and false if not found.
func (i *I18n) TryMatchString(s string) (language.Tag, int, bool) {
	if tag, err := language.Parse(s); err == nil {
		if tag, index, conf := i.matcher.Match(tag); conf > language.Low {
			return tag, index, true
		}
	}

	return language.Und, -1, false
}

// Tr is package-level function which calls the `Default.Tr` method.
//
// See `I18n#Tr` method for more.
func Tr(lang, format string, args ...interface{}) string {
	return Default.Tr(lang, format, args...)
}

// Tr returns a translated message based on the "lang" language code
// and its key(format) with any optional arguments attached to it.
//
// It returns an empty string if "format" not matched.
func (i *I18n) Tr(lang, format string, args ...interface{}) string {
	_, index, ok := i.TryMatchString(lang)
	if !ok {
		index = 0
	}

	loc := i.localizer.GetLocale(index)
	if loc != nil {
		return loc.GetMessage(format, args...)
	}

	return fmt.Sprintf(format, args...)
}

const acceptLanguageHeaderKey = "Accept-Language"

// GetLocale is package-level function which calls the `Default.GetLocale` method.
//
// See `I18n#GetLocale` method for more.
func GetLocale(r *http.Request) Locale {
	return Default.GetLocale(r)
}

// GetLocale returns the found locale of a request.
// It will return the first registered language if nothing else matched.
func (i *I18n) GetLocale(r *http.Request) Locale {
	var (
		index int
		ok    bool
	)

	if i.ContextKey != nil {
		if v := r.Context().Value(i.ContextKey); v != nil {
			if s, isString := v.(string); isString {
				_, index, ok = i.TryMatchString(s)
			}
		}
	}

	if !ok && i.ExtractFunc != nil {
		if v := i.ExtractFunc(r); v != "" {
			_, index, ok = i.TryMatchString(v)

		}
	}

	if !ok && i.URLParameter != "" {
		if v := r.URL.Query().Get(i.URLParameter); v != "" {
			_, index, ok = i.TryMatchString(v)
		}
	}

	if !ok && i.Cookie != "" {
		cookie, err := r.Cookie(i.Cookie)
		if err == nil {
			_, index, ok = i.TryMatchString(cookie.Value) // url.QueryUnescape(cookie.Value)
		}
	}

	if !ok && i.Subdomain {
		if v, _ := getSubdomain(r); v != "" {
			_, index, ok = i.TryMatchString(v)
		}
	}

	if !ok {
		if v := r.Header.Get(acceptLanguageHeaderKey); v != "" {
			desired, _, err := language.ParseAcceptLanguage(v)
			if err == nil {
				if _, idx, conf := i.matcher.Match(desired...); conf > language.Low {
					index = idx
				}
			}
		}
	}

	// if 0 then it defaults to the first language.
	return i.localizer.GetLocale(index)
}

// GetMessage is package-level function which calls the `Default.GetMessage` method.
//
// See `I18n#GetMessage` method for more.
func GetMessage(r *http.Request, format string, args ...interface{}) string {
	return Default.GetMessage(r, format, args...)
}

// GetMessage returns the localized text message for this "r" request based on the key "format".
func (i *I18n) GetMessage(r *http.Request, format string, args ...interface{}) string {
	loc := i.GetLocale(r)
	if loc != nil {
		return loc.GetMessage(format, args...)
	}

	return fmt.Sprintf(format, args...)
}

// Router is package-level function which calls the `Default.Router` method.
//
// See `I18n#Router` method for more.
func Router(next http.Handler) http.Handler {
	return Default.Router(next)
}

// Router returns a new router wrapper.
// It compares the path prefix for translated language and
// local redirects the requested path with the selected (from the path) language to the router.
func (i *I18n) Router(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		found := false
		path := r.URL.Path[1:]

		if idx := strings.IndexByte(path, '/'); idx > 0 {
			path = path[:idx]
		}

		if path != "" {
			if tag, _, ok := i.TryMatchString(path); ok {
				lang := tag.String()

				path = r.URL.Path[len(path)+1:]
				if path == "" {
					path = "/"
				}

				r.RequestURI = path
				r.URL.Path = path

				if i.ContextKey != nil {
					r = r.WithContext(context.WithValue(r.Context(), i.ContextKey, lang))
				}
				r.Header.Set(acceptLanguageHeaderKey, lang)
				found = true
			}
		}

		if !found && i.Subdomain {
			if subdomain, host := getSubdomain(r); subdomain != "" {
				if tag, _, ok := i.TryMatchString(subdomain); ok {
					lang := tag.String()

					r.URL.Host = host
					r.Host = host

					if i.ContextKey != nil {
						r = r.WithContext(context.WithValue(r.Context(), i.ContextKey, lang))
					}
					r.Header.Set(acceptLanguageHeaderKey, lang)
				}
			}
		}

		next.ServeHTTP(w, r)
	})

}

func getSubdomain(r *http.Request) (subdomain, host string) {
	host = r.Host
	if host == "" {
		host = r.URL.Host
	}

	if index := strings.IndexByte(host, '.'); index > 0 {
		if subdomain = host[0:index]; subdomain != "" {
			host = host[index+1:]
		}
	}

	return
}
