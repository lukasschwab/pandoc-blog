// pandoc-blog is a static site generator that uses pandoc to convert
// Markdown posts into HTML pages, generates an index page, and produces
// a JSON Feed.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	postsDir      = "posts"
	genDir        = "gen"
	templatesDir  = "templates"
	postTemplate  = "templates/post.html"
	indexTemplate = "templates/index.html"
	indexMD       = "index.md"
	indexHTML     = "index.html"
	feedFile      = "feed.json"
)

// frontmatter represents the YAML front matter in a Markdown post.
type frontmatter struct {
	Title    string    `yaml:"title"`
	Author   string    `yaml:"author"`
	Date     time.Time `yaml:"date"`
	Abstract string    `yaml:"abstract"`
	Draft    bool      `yaml:"draft"`
}

// postMeta holds parsed metadata plus the source filename.
type postMeta struct {
	frontmatter
	Filename string // e.g. "example-post.md"
}

// staticPath returns the generated HTML path for this post,
// prefixed with "./" for use in URLs.
func (p postMeta) staticPath() string {
	base := strings.TrimSuffix(p.Filename, ".md")
	return "./" + genDir + "/" + base + ".html"
}

// parseFrontmatter extracts YAML front matter from a Markdown file.
// It expects the file to start with "---\n" and end the front matter
// block with another "---\n".
func parseFrontmatter(path string) (frontmatter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return frontmatter{}, err
	}

	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return frontmatter{}, fmt.Errorf("%s: no front matter found", path)
	}

	// Find the closing ---.
	end := strings.Index(content[3:], "\n---")
	if end == -1 {
		return frontmatter{}, fmt.Errorf("%s: unclosed front matter", path)
	}

	yamlBlock := content[4 : 3+end] // skip the opening "---\n"

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return frontmatter{}, fmt.Errorf("%s: %w", path, err)
	}

	// If the parsed date has no timezone info (zero location), assume UTC.
	if fm.Date.Location() == time.UTC {
		// Already UTC, fine.
	} else if fm.Date.IsZero() {
		log.Printf("[WARN] no date for post %s; using now", path)
		fm.Date = time.Now().UTC()
	}

	return fm, nil
}

// loadPosts reads all .md files from the posts directory.
func loadPosts() ([]postMeta, error) {
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}

	var posts []postMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		fm, err := parseFrontmatter(filepath.Join(postsDir, e.Name()))
		if err != nil {
			log.Printf("[WARN] skipping %s: %v", e.Name(), err)
			continue
		}
		if fm.Title == "" {
			fm.Title = strings.TrimSuffix(e.Name(), ".md")
		}
		posts = append(posts, postMeta{
			frontmatter: fm,
			Filename:    e.Name(),
		})
	}
	return posts, nil
}

// runPandoc shells out to pandoc with the given arguments.
func runPandoc(args ...string) error {
	cmd := exec.Command("pandoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// buildPost converts a single Markdown post to HTML via pandoc.
func buildPost(filename string) error {
	src := filepath.Join(postsDir, filename)
	dst := filepath.Join(genDir, strings.TrimSuffix(filename, ".md")+".html")
	return runPandoc(
		"-f", "markdown+fenced_divs",
		"-s", src,
		"-o", dst,
		"--template", postTemplate,
		"--css=../styles/common.css",
	)
}

// generateIndexMD writes the intermediate index.md from post metadata.
func generateIndexMD(posts []postMeta) error {
	var buf bytes.Buffer
	if len(posts) == 0 {
		buf.WriteString("There aren't any posts yet.\n")
	} else {
		for _, p := range posts {
			fmt.Fprintf(&buf, "## [%s](%s)\n", p.Title, p.staticPath())
			if p.Abstract != "" {
				fmt.Fprintf(&buf, "%s &middot; %s\n",
					p.Date.Format("January 02, 2006"),
					p.Abstract,
				)
			}
			buf.WriteString("\n")
		}
	}
	return os.WriteFile(indexMD, buf.Bytes(), 0644)
}

// buildIndex runs pandoc to convert index.md to index.html, then cleans up.
func buildIndex() error {
	err := runPandoc(
		"-s", indexMD,
		"-o", indexHTML,
		"--template", indexTemplate,
		"--css=./styles/common.css",
		"--css=./styles/index.css",
	)
	os.Remove(indexMD) // clean up intermediate file
	return err
}

// --- JSON Feed types (jsonfeed.org/version/1.1) ---

type jsonFeed struct {
	Version string         `json:"version"`
	Title   string         `json:"title"`
	Expired bool           `json:"expired"`
	Items   []jsonFeedItem `json:"items"`
}

type jsonFeedItem struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	ContentHTML   string `json:"content_html,omitempty"`
	Summary       string `json:"summary,omitempty"`
	DatePublished string `json:"date_published,omitempty"`
}

// generateFeed writes feed.json.
func generateFeed(posts []postMeta) error {
	feed := jsonFeed{
		Version: "https://jsonfeed.org/version/1.1",
		Title:   "blog",
		Expired: false,
	}

	for _, p := range posts {
		url := p.staticPath()

		item := jsonFeedItem{
			ID:    url,
			URL:   url,
			Title: p.Title,
		}

		if !p.Date.IsZero() {
			item.DatePublished = p.Date.Format(time.RFC3339)
		}
		if p.Abstract != "" {
			item.Summary = p.Abstract
		}

		// Read the generated HTML to embed in the feed.
		htmlPath := p.staticPath()
		if data, err := os.ReadFile(htmlPath); err == nil {
			item.ContentHTML = string(data)
		}

		feed.Items = append(feed.Items, item)
	}

	out, err := json.MarshalIndent(feed, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(feedFile, out, 0644)
}

func main() {
	log.SetFlags(0)

	// Ensure gen/ exists.
	if err := os.MkdirAll(genDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Load post metadata.
	posts, err := loadPosts()
	if err != nil {
		log.Fatal(err)
	}

	// Build each post.
	for _, p := range posts {
		log.Printf("building %s", p.Filename)
		if err := buildPost(p.Filename); err != nil {
			log.Fatal(err)
		}
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

	// Generate index.
	log.Println("generating index")
	if err := generateIndexMD(published); err != nil {
		log.Fatal(err)
	}
	if err := buildIndex(); err != nil {
		log.Fatal(err)
	}

	// Generate JSON feed.
	log.Println("generating feed.json")
	if err := generateFeed(published); err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}
