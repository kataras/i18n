// Package i18n provides internalization and localization features.
package i18n

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/kataras/i18n/internal"

	"golang.org/x/net/publicsuffix"
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
	// Map is just an alias of the map[string]interface{} type.
	Map = map[string]interface{}

	// Locale is the type which the `Localizer.GetLocale` method returns.
	// It serves the translations based on "key" or format. See its `GetMessage`.
	Locale = internal.Locale

	// MessageFunc is the function type to modify the behavior when a key or language was not found.
	// All language inputs fallback to the default locale if not matched.
	// This is why this signature accepts both input and matched languages, so caller
	// can provide better messages.
	//
	// The first parameter is set to the client real input of the language,
	// the second one is set to the matched language (default one if input wasn't matched)
	// and the third and forth are the translation format/key and its optional arguments.
	MessageFunc = internal.MessageFunc

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
		GetLocale(index int) *Locale
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
	// DefaultMessageFunc is the field which can be used
	// to modify the behavior when a key or language was not found.
	// All language inputs fallback to the default locale if not matched.
	// This is why this one accepts both input and matched languages,
	// so the caller can be more expressful knowing those.
	//
	// Defaults to nil.
	DefaultMessageFunc MessageFunc
	// ExtractFunc is the type signature for declaring custom logic
	// to extract the language tag name.
	ExtractFunc func(*http.Request) string
	// If not empty, it is language identifier by url query.
	URLParameter string
	// If not empty, it is language identifier by cookie of this name.
	Cookie string
	// If true then a subdomain can be a language identifier too.
	Subdomain bool
	// If true then it will return empty string when translation for a a specific language's key was not found.
	// Defaults to false, fallback defaultLang:key will be used.
	Strict bool
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

	i := new(I18n)
	i.loader = loader
	i.matcher = &Matcher{
		strict:             len(tags) > 0,
		Languages:          tags,
		matcher:            language.NewMatcher(tags),
		defaultMessageFunc: i.DefaultMessageFunc,
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
	// defaultMessageFunc passed by the i18n structure.
	defaultMessageFunc MessageFunc
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

func reverseStrings(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func parseLanguage(path string) (language.Tag, bool) {
	if idx := strings.LastIndexByte(path, '.'); idx > 0 {
		path = path[0:idx]
	}

	// path = strings.ReplaceAll(path, "..", "")

	names := strings.FieldsFunc(path, func(r rune) bool {
		return r == '_' || r == os.PathSeparator || r == '/' || r == '.'
	})

	names = reverseStrings(names) // see https://github.com/kataras/i18n/issues/1

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
// It returns an empty string if "lang" not matched, unless DefaultMessageFunc.
// It returns the default language's translation if "key" not matched, unless DefaultMessageFunc.
func (i *I18n) Tr(lang, format string, args ...interface{}) (msg string) {
	_, index, ok := i.TryMatchString(lang)
	if !ok {
		index = 0
	}

	langMatched := ""

	loc := i.localizer.GetLocale(index)
	if loc != nil {
		langMatched = loc.Language()

		msg = loc.GetMessage(format, args...)
		if msg == "" && i.DefaultMessageFunc == nil && !i.Strict && index > 0 {
			// it's not the default/fallback language and not message found for that lang:key.
			msg = i.localizer.GetLocale(0).GetMessage(format, args...)
		}
	}

	if msg == "" && i.DefaultMessageFunc != nil {
		msg = i.DefaultMessageFunc(lang, langMatched, format, args...)
	}

	return
}

const acceptLanguageHeaderKey = "Accept-Language"

// GetLocale is package-level function which calls the `Default.GetLocale` method.
//
// See `I18n#GetLocale` method for more.
func GetLocale(r *http.Request) *Locale {
	return Default.GetLocale(r)
}

// GetLocale returns the found locale of a request.
// It will return the first registered language if nothing else matched.
func (i *I18n) GetLocale(r *http.Request) *Locale {
	var (
		index int
		ok    bool
	)

	if i.ContextKey != nil {
		if v := r.Context().Value(i.ContextKey); v != nil {
			if s, isString := v.(string); isString {
				if v == "default" {
					index = 0 // no need to call `TryMatchString` and spend time.
				} else {
					_, index, _ = i.TryMatchString(s)
				}

				locale := i.localizer.GetLocale(index)
				if locale == nil {
					return nil
				}

				return locale
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

	// if index == 0 then it defaults to the first language.
	locale := i.localizer.GetLocale(index)
	if locale == nil {
		return nil
	}

	return locale
}

// GetMessage is package-level function which calls the `Default.GetMessage` method.
//
// See `I18n#GetMessage` method for more.
func GetMessage(r *http.Request, format string, args ...interface{}) string {
	return Default.GetMessage(r, format, args...)
}

// GetMessage returns the localized text message for this "r" request based on the key "format".
// It returns an empty string if locale or format not found.
func (i *I18n) GetMessage(r *http.Request, format string, args ...interface{}) (msg string) {
	loc := i.GetLocale(r)
	langMatched := ""
	if loc != nil {
		langMatched = loc.Language()
		// it's not the default/fallback language and not message found for that lang:key.
		msg = loc.GetMessage(format, args...)
		if msg == "" && i.DefaultMessageFunc == nil && !i.Strict && loc.Index() > 0 {
			return i.localizer.GetLocale(0).GetMessage(format, args...)
		}
	}

	if msg == "" && i.DefaultMessageFunc != nil && i.ContextKey != nil {
		if v := r.Context().Value(i.ContextKey); v != nil {
			if langInput, ok := v.(string); ok {
				msg = i.DefaultMessageFunc(langInput, langMatched, format, args...)
			}
		}
	}

	return
}

// Router is package-level function which calls the `Default.Router` method.
//
// See `I18n#Router` method for more.
func Router(next http.Handler) http.Handler {
	return Default.Router(next)
}

func (i *I18n) setLang(w http.ResponseWriter, r *http.Request, lang string) {
	if i.Cookie != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  i.Cookie,
			Value: lang,
			// allow subdomain sharing.
			Domain:   getDomain(getHost(r)),
			SameSite: http.SameSiteLaxMode,
		})
	} else if i.URLParameter != "" {
		r.URL.Query().Set(i.URLParameter, lang)
	}

	r.Header.Set(acceptLanguageHeaderKey, lang)
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
				i.setLang(w, r, lang)
				found = true
			}
		}

		if !found && i.Subdomain {
			host := getHost(r)
			if dotIdx := strings.IndexByte(host, '.'); dotIdx > 0 {
				if subdomain := host[0:dotIdx]; subdomain != "" {
					if tag, _, ok := i.TryMatchString(subdomain); ok {
						host = host[dotIdx+1:]
						r.URL.Host = host
						r.Host = host
						i.setLang(w, r, tag.String())
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getHost(r *http.Request) string {
	// contains subdomain.
	if host := r.URL.Host; host != "" {
		return host
	}
	return r.Host
}

// GetDomain resolves and returns the server's domain.
func getDomain(hostport string) string {
	host := hostport
	if tmp, _, err := net.SplitHostPort(hostport); err == nil {
		host = tmp
	}

	switch host {
	// We could use the netutil.LoopbackRegex but leave it as it's for now, it's faster.
	case "localhost", "127.0.0.1", "0.0.0.0", "::1", "[::1]", "0:0:0:0:0:0:0:0", "0:0:0:0:0:0:0:1":
		// loopback.
		return "localhost"
	default:
		if domain, err := publicsuffix.EffectiveTLDPlusOne(host); err == nil {
			host = domain
		}

		return host
	}
}

func getSubdomain(r *http.Request) (subdomain, host string) {
	host = getHost(r)

	if index := strings.IndexByte(host, '.'); index > 0 {
		if subdomain = host[0:index]; subdomain != "" {
			host = host[index+1:]
		}
	}

	return
}
