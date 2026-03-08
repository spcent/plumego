// Package docsite serves markdown documentation as HTML pages.
package docsite

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// FooterLink is a navigation link rendered in the page footer.
type FooterLink struct {
	Href  string
	Label string
}

// DocPage represents a single markdown document in the navigation tree.
type DocPage struct {
	Lang  string
	Slug  string
	Title string
}

// docTemplateData is the template payload for the docs HTML template.
type docTemplateData struct {
	Title        string
	Lang         string
	CurrentSlug  string
	Content      template.HTML
	Navigation   map[string][]DocPage
	LanguageList []string
	FooterLinks  []FooterLink
}

// Site renders markdown documentation as HTML.
type Site struct {
	fs          fs.FS
	nav         map[string][]DocPage
	defaultLang string
	tmpl        *template.Template
	footerLinks []FooterLink
}

// New creates a new documentation site rooted at the first discoverable docs
// directory, using staticFS to parse the docs HTML template.
// footerLinks are rendered in the page footer on every docs page.
func New(staticFS fs.FS, footerLinks []FooterLink) (*Site, error) {
	docsPath := locateDocsPath()
	if docsPath == "" {
		return nil, fmt.Errorf("docs directory not found")
	}

	f := os.DirFS(docsPath)
	nav, err := buildNav(f)
	if err != nil {
		return nil, fmt.Errorf("build docs nav: %w", err)
	}

	defaultLang := "en"
	if _, ok := nav[defaultLang]; !ok {
		for lang := range nav {
			defaultLang = lang
			break
		}
	}

	tmpl, err := template.ParseFS(staticFS,
		"templates/docs.html",
		"templates/partials.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse docs template: %w", err)
	}

	return &Site{
		fs:          f,
		nav:         nav,
		defaultLang: defaultLang,
		tmpl:        tmpl,
		footerLinks: footerLinks,
	}, nil
}

// Handler returns the HTTP handler for all documentation routes.
func (s *Site) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := s.defaultLang
		slug := ""
		trimmed := strings.TrimPrefix(r.URL.Path, "/docs")
		trimmed = strings.Trim(trimmed, "/")
		if trimmed != "" {
			parts := strings.Split(trimmed, "/")
			if len(parts) > 0 && parts[0] != "" {
				lang = parts[0]
			}
			if len(parts) > 1 {
				slug = parts[1]
			}
		}
		if slug == "" {
			s.renderIndex(w, lang)
			return
		}
		s.renderDoc(w, r, lang, slug)
	}
}

func (s *Site) renderIndex(w http.ResponseWriter, lang string) {
	content := markdownToHTML("# Documentation\nSelect a language and module on the left to start reading, or directly access /docs/{lang}/{module}.\n\nUse the navigation on the left to browse modules by language.")
	s.renderPage(w, docTemplateData{
		Title:        "Plumego Docs",
		Lang:         lang,
		CurrentSlug:  "",
		Content:      content,
		Navigation:   s.nav,
		LanguageList: s.languages(),
		FooterLinks:  s.footerLinks,
	})
}

func (s *Site) renderDoc(w http.ResponseWriter, r *http.Request, lang, slug string) {
	filePath := path.Join(lang, slug+".md")
	data, err := fs.ReadFile(s.fs, filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	content := markdownToHTML(string(data))
	pageTitle := slug
	if pages, ok := s.nav[lang]; ok {
		for _, p := range pages {
			if p.Slug == slug {
				pageTitle = p.Title
				break
			}
		}
	}
	s.renderPage(w, docTemplateData{
		Title:        pageTitle,
		Lang:         lang,
		CurrentSlug:  slug,
		Content:      content,
		Navigation:   s.nav,
		LanguageList: s.languages(),
		FooterLinks:  s.footerLinks,
	})
}

func (s *Site) languages() []string {
	langs := make([]string, 0, len(s.nav))
	for lang := range s.nav {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

func (s *Site) renderPage(w http.ResponseWriter, data docTemplateData) {
	buf := &bytes.Buffer{}
	if err := s.tmpl.ExecuteTemplate(buf, "docs.html", data); err != nil {
		http.Error(w, fmt.Sprintf("render docs: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func locateDocsPath() string {
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "examples", "docs"),
			filepath.Join(cwd, "docs"),
			filepath.Join(filepath.Dir(cwd), "docs"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "examples", "docs"),
		)
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

func buildNav(f fs.FS) (map[string][]DocPage, error) {
	nav := make(map[string][]DocPage)
	langs, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, err
	}
	for _, langEntry := range langs {
		if !langEntry.IsDir() {
			continue
		}
		lang := langEntry.Name()
		files, err := fs.ReadDir(f, lang)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
				continue
			}
			slug := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			title := extractTitle(f, path.Join(lang, file.Name()))
			nav[lang] = append(nav[lang], DocPage{Lang: lang, Slug: slug, Title: title})
		}
		sort.Slice(nav[lang], func(i, j int) bool { return nav[lang][i].Title < nav[lang][j].Title })
	}
	return nav, nil
}

func extractTitle(f fs.FS, filePath string) string {
	file, err := f.Open(filePath)
	if err != nil {
		return filePath
	}
	defer file.Close()

	buf := make([]byte, 1024*1024)
	n, _ := file.Read(buf)
	for _, line := range strings.Split(string(buf[:n]), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimLeft(line, "# ")
		}
	}
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}
