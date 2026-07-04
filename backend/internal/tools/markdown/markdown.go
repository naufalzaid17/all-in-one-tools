// Package markdown implements the Markdown Toolkit tool: merging
// multiple .md files into a single master context document (for AI
// agent consumption) and converting Markdown to HTML via goldmark.
//
// Markdown files are small text documents, so per the application's
// memory strategy they are processed entirely in memory with
// bytes.Buffer — no temp files involved. Upload limits are tightened
// accordingly.
package markdown

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/fileio"
)

const (
	contentTypeMarkdown = "text/markdown; charset=utf-8"
	contentTypeHTML     = "text/html; charset=utf-8"
	// Markdown is small text; keep uploads well under the global caps.
	maxMarkdownFile  = 10 << 20 // 10 MiB per file
	maxMarkdownTotal = 40 << 20 // 40 MiB per request
)

var uploadOpts = fileio.ParseOptions{
	MaxFileSize:       maxMarkdownFile,
	MaxTotalSize:      maxMarkdownTotal,
	AllowedExtensions: []string{".md", ".markdown", ".txt"},
}

// converter renders GitHub-flavoured Markdown. XHTML/unsafe raw HTML is
// left enabled deliberately: the output is a downloaded document owned
// by the user, not something this app serves and executes.
var converter = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

// Tool is the Markdown Toolkit tool module.
type Tool struct{}

// New returns the Markdown Toolkit tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "markdown" }
func (t *Tool) Name() string            { return "Markdown Toolkit" }
func (t *Tool) Category() core.Category { return core.CategoryDocument }
func (t *Tool) Description() string {
	return "Merge .md files into one master context file and convert Markdown to HTML."
}

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/markdown.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/merge", t.handleMerge)
	r.Post("/to-html", t.handleToHTML)
}

// handleMerge concatenates the uploaded Markdown files into a single
// master document. Each source is delimited with an HTML comment and a
// heading so both humans and AI agents can locate section boundaries.
func (t *Tool) handleMerge(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := fileio.ParseRequest(w, r, uploadOpts)
	if !ok {
		return
	}
	defer ws.Close()

	if len(form.Files) < 2 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "merging requires at least two Markdown files")
		return
	}

	var buf bytes.Buffer
	buf.WriteString("<!-- Master context file: merged by All-in-One Tools -->\n\n")

	for i, f := range form.Files {
		// Small text file: safe to load into memory (limits enforced
		// during upload).
		data, err := os.ReadFile(f.Path)
		if err != nil {
			core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not read "+f.Name)
			return
		}
		if i > 0 {
			buf.WriteString("\n\n---\n\n")
		}
		fmt.Fprintf(&buf, "<!-- ===== Begin: %s ===== -->\n\n", f.Name)
		buf.Write(normalizeTrailingNewline(data))
		fmt.Fprintf(&buf, "\n<!-- ===== End: %s ===== -->\n", f.Name)
	}

	core.RespondBytesDownload(w, buf.Bytes(), "master-context.md", contentTypeMarkdown)
}

// handleToHTML converts a single Markdown file to HTML. With
// standalone=true (default) the output is a complete, minimally styled
// HTML document; standalone=false yields just the rendered fragment.
func (t *Tool) handleToHTML(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := fileio.ParseRequest(w, r, uploadOpts)
	if !ok {
		return
	}
	defer ws.Close()

	if len(form.Files) != 1 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "exactly one Markdown file is required")
		return
	}
	in := form.Files[0]

	source, err := os.ReadFile(in.Path)
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not read "+in.Name)
		return
	}

	var body bytes.Buffer
	if err := converter.Convert(source, &body); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "conversion failed: "+err.Error())
		return
	}

	out := body.Bytes()
	if form.Value("standalone") != "false" {
		title := strings.TrimSuffix(in.Name, ".md")
		out = wrapDocument(title, out)
	}

	name := deriveHTMLName(in.Name)
	core.RespondBytesDownload(w, out, name, contentTypeHTML)
}

func normalizeTrailingNewline(data []byte) []byte {
	return append(bytes.TrimRight(data, "\r\n"), '\n')
}

func deriveHTMLName(original string) string {
	base := original
	for _, ext := range []string{".md", ".markdown", ".txt"} {
		if strings.HasSuffix(strings.ToLower(base), ext) {
			base = base[:len(base)-len(ext)]
			break
		}
	}
	if base == "" {
		base = "converted"
	}
	return base + ".html"
}

// wrapDocument embeds the rendered fragment in a self-contained HTML
// page with typography that mirrors GitHub's rendering.
func wrapDocument(title string, fragment []byte) []byte {
	var doc bytes.Buffer
	doc.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	doc.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	fmt.Fprintf(&doc, "<title>%s</title>\n", html.EscapeString(title))
	doc.WriteString(`<style>
  body { max-width: 860px; margin: 2rem auto; padding: 0 1.25rem;
         font: 16px/1.6 -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
         color: #1f2328; }
  pre { background: #f6f8fa; padding: 1rem; border-radius: 6px; overflow-x: auto; }
  code { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 85%; }
  :not(pre) > code { background: #f6f8fa; padding: .2em .4em; border-radius: 6px; }
  table { border-collapse: collapse; display: block; overflow-x: auto; }
  th, td { border: 1px solid #d1d9e0; padding: 6px 13px; }
  blockquote { margin: 0; padding-left: 1em; color: #59636e; border-left: .25em solid #d1d9e0; }
  img { max-width: 100%; }
  @media (prefers-color-scheme: dark) {
    body { background: #0d1117; color: #f0f6fc; }
    pre, :not(pre) > code { background: #151b23; }
    th, td { border-color: #3d444d; }
    blockquote { color: #9198a1; border-left-color: #3d444d; }
  }
</style>
</head>
<body>
`)
	doc.Write(fragment)
	doc.WriteString("</body>\n</html>\n")
	return doc.Bytes()
}
