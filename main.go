// pandoc-blog is a static site generator that uses pandoc to convert
// Markdown posts into HTML pages, generates an index page, and produces
// a JSON Feed.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/kelseyhightower/envconfig"
	jsonfeed "github.com/lukasschwab/go-jsonfeed"
)

// config holds all configurable paths and values. Defaults can be
// overridden via environment variables prefixed with BLOG_, e.g.
// BLOG_POSTS_DIR, BLOG_FEED_TITLE.
type config struct {
	// PostsDir is the directory containing Markdown source posts.
	PostsDir string `envconfig:"POSTS_DIR" default:"posts"`
	// GenDir is the directory for generated per-post HTML files.
	GenDir string `envconfig:"GEN_DIR" default:"gen"`
	// IndexHTML is the generated index page.
	IndexHTML string `envconfig:"INDEX_HTML" default:"index.html"`
	// IndexTemplate is the pandoc template for the index page.
	IndexTemplate string `envconfig:"INDEX_TEMPLATE" default:"templates/index.html"`
	// FeedFile is the generated JSON Feed file.
	FeedFile string `envconfig:"FEED_FILE" default:"feed.json"`
	// FeedTitle is the title of the JSON Feed.
	FeedTitle string `envconfig:"FEED_TITLE" default:"blog"`
	// Domain is the base URL of the blog, used for feed metadata.
	// When set, home_page_url and feed_url are included in the JSON feed,
	// and item URLs are absolute. When empty, item URLs are relative.
	Domain string `envconfig:"DOMAIN" default:""`
}

// postFrontmatter represents the YAML front matter in a Markdown post.
type postFrontmatter struct {
	Title    string    `yaml:"title"`
	Author   string    `yaml:"author"`
	Date     time.Time `yaml:"date"`
	Abstract string    `yaml:"abstract"`
	Draft    bool      `yaml:"draft"`
}

// postMeta holds parsed front matter metadata plus the source filename.
type postMeta struct {
	postFrontmatter
	Filename string // e.g. "example-post.md"
}

// staticPath returns the generated HTML path for this post,
// prefixed with "./" for use in URLs and relative references.
func (p postMeta) staticPath(genDir string) string {
	base := strings.TrimSuffix(p.Filename, ".md")
	return "./" + genDir + "/" + base + ".html"
}

// parsePost reads a Markdown file and extracts its YAML front matter
// using the adrg/frontmatter library. If no date is provided, the
// current time in UTC is used as a fallback.
func parsePost(path string) (postFrontmatter, error) {
	f, err := os.Open(path)
	if err != nil {
		return postFrontmatter{}, err
	}
	defer f.Close()

	var fm postFrontmatter
	if _, err := frontmatter.Parse(f, &fm); err != nil {
		return postFrontmatter{}, fmt.Errorf("%s: %w", path, err)
	}

	if fm.Date.IsZero() {
		log.Printf("[WARN] no date for post %s; using now", path)
		fm.Date = time.Now().UTC()
	}

	return fm, nil
}

// loadPosts reads all .md files from the posts directory and returns
// their parsed metadata. Files that fail to parse are logged and skipped.
func loadPosts(postsDir string) ([]postMeta, error) {
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}

	var posts []postMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		fm, err := parsePost(filepath.Join(postsDir, e.Name()))
		if err != nil {
			log.Printf("[WARN] skipping %s: %v", e.Name(), err)
			continue
		}
		if fm.Title == "" {
			fm.Title = strings.TrimSuffix(e.Name(), ".md")
		}
		posts = append(posts, postMeta{
			postFrontmatter: fm,
			Filename:        e.Name(),
		})
	}
	return posts, nil
}

// indexMarkdown generates index content as Markdown from post metadata.
// Each non-draft post is rendered as a Markdown heading linking to its
// generated HTML, with an optional date and abstract line. Using a
// Markdown intermediate allows pandoc markdown in titles and abstracts.
func indexMarkdown(posts []postMeta, genDir string) []byte {
	var buf bytes.Buffer
	if len(posts) == 0 {
		buf.WriteString("There aren't any posts yet.\n")
	} else {
		for _, p := range posts {
			fmt.Fprintf(&buf, "## [%s](%s)\n", p.Title, p.staticPath(genDir))
			if p.Abstract != "" {
				fmt.Fprintf(&buf, "%s &middot; %s\n",
					p.Date.Format("January 02, 2006"),
					p.Abstract,
				)
			}
			buf.WriteString("\n")
		}
	}
	return buf.Bytes()
}

// buildIndex pipes index Markdown to pandoc via stdin to produce
// index.html, avoiding any intermediate files on disk.
func buildIndex(cfg config, markdown []byte) error {
	cmd := exec.Command("pandoc",
		"-s",
		"-f", "markdown",
		"-o", cfg.IndexHTML,
		"--metadata", "title=index",
		"--template", cfg.IndexTemplate,
		"--css=./styles/common.css",
		"--css=./styles/index.css",
	)
	cmd.Stdin = bytes.NewReader(markdown)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc (index): %w", err)
	}
	return nil
}

// generateFeed writes feed.json using the go-jsonfeed library.
// Each published post becomes a feed item with its full generated HTML
// embedded as content_html.
func generateFeed(cfg config, posts []postMeta) error {
	var items []jsonfeed.Item
	for _, p := range posts {
		url := p.staticPath(cfg.GenDir)
		if cfg.Domain != "" {
			url = strings.TrimRight(cfg.Domain, "/") + strings.TrimPrefix(url, ".")
		}

		item := jsonfeed.NewItem(url)
		item.URL = new(url)
		item.Title = new(p.Title)

		if !p.Date.IsZero() {
			item.DatePublished = new(p.Date.Format(time.RFC3339))
		}
		if p.Abstract != "" {
			item.Summary = new(p.Abstract)
		}

		// Read the generated HTML to embed in the feed.
		// NOTE: relative links (incl. img sources) won't work in a feed
		// reader, but this is a better best effort than just including
		// abstracts.
		htmlPath := filepath.Join(cfg.GenDir, strings.TrimSuffix(p.Filename, ".md")+".html")
		if data, err := os.ReadFile(htmlPath); err == nil {
			item.ContentHTML = new(string(data))
		}

		items = append(items, item)
	}

	feed := jsonfeed.NewFeed(cfg.FeedTitle, items)
	if cfg.Domain != "" {
		feed.HomePageURL = new(cfg.Domain)
		feed.FeedURL = new(strings.TrimRight(cfg.Domain, "/") + "/" + cfg.FeedFile)
	}
	feed.Expired = new(false)

	data, err := feed.ToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.FeedFile, data, 0644)
}

func main() {
	log.SetFlags(0)

	var cfg config
	if err := envconfig.Process("blog", &cfg); err != nil {
		log.Fatal(err)
	}

	// Ensure gen/ exists.
	if err := os.MkdirAll(cfg.GenDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Load post metadata.
	posts, err := loadPosts(cfg.PostsDir)
	if err != nil {
		log.Fatal(err)
	}

	// Filter drafts and sort by date descending for index/feed.
	var published []postMeta
	for _, p := range posts {
		if !p.Draft {
			published = append(published, p)
		}
	}
	sort.Slice(published, func(i, j int) bool {
		return published[i].Date.After(published[j].Date)
	})

	// Generate index by piping Markdown to pandoc.
	log.Println("generating index")
	if err := buildIndex(cfg, indexMarkdown(published, cfg.GenDir)); err != nil {
		log.Fatal(err)
	}

	// Generate JSON feed (post HTML files must already exist in gen/).
	log.Println("generating feed.json")
	if err := generateFeed(cfg, published); err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}
