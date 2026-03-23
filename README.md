# pandoc-blog

This is a *very* basic Pandoc static site generator.

I've avoided writing CSS beyond some very basic readability improvements/demonstrating that CSS can be added. The goal here is just to populate well-structured HTML documents and a generated JSON feed. [View a demo site.](http://lukasschwab.me/pandoc-blog/index.html)

Deeply unattractive out of the box? Yes. Easy to customize? I hope so.

## Requirements

+ `pandoc`
+ Go 1.22+

## Usage

1. Clone this repository.
2. Write Pandoc-compatible Markdown files in `posts`. These should include YAML frontmatter for generating the index:
    + `title`: a human-readable title for this post.
    + `date`: an ISO 8601 date (`make date`).
    + `abstract`: a summary you want to appear on the index. This can include valid Pandoc markdown.
    + `draft`: set to `true` to exclude from the index and feed.
3. Run `make all` to build an HTML file for each Markdown page, generate `index.html`, and produce `feed.json`.

### Utilities

+ `make build`: compile the Go binary.
+ `make clean`: remove generated files and the binary.
+ `make hook`: configure a git hook to run `make all` before each commit (so each commit contains an up-to-date static site).

## How it works

`main.go` is a single Go program that replaces the old `make_index.py` + `Makefile` pipeline:

1. It reads all `.md` files in `posts/`, parses their YAML frontmatter.
2. It shells out to `pandoc` to convert each post to HTML in `gen/`, using `templates/post.html` and `styles/common.css`.
3. It generates an intermediate `index.md` with links and metadata for all non-draft posts (sorted newest-first), then runs `pandoc` to produce `index.html` using `templates/index.html`.
4. It generates `feed.json` (JSON Feed 1.1) with the full HTML content of each post.

## Customization

A general rule of thumb: changes to the HTML are predictable; changes to pre-`pandoc` Markdown are unpredictable. Markdown intermediates are antipatterns.

+ Want to change how posts are represented in the index?<br>Modify `main.go` (`generateIndexMD`).

+ Want to add static elements, e.g. a section with "about me" info or social links?<br>Modify `templates/index.html` to only change the index.<br>Modify `templates/post.html` to only change the post pages.

+ Want to change how the index is styled?<br>Modify `styles/index.css`.

+ Want to change how the whole generated site is styled?<br>Modify `styles/common.css`.

### Fenced `div`s

`pandoc-blog` converts Markdown to HTML with [`pandoc`'s `fenced_divs` extension](https://pandoc.org/MANUAL.html#extension-fenced_divs) enabled. You can use this to define a `div`––complete with HTML attributes––in your markup:

```markdown
This text is outside of the fenced `div`.

::: {.addendum date="Oct. 12, 2020"}

This text is in a div with class `addendum` and attribute `date="Oct. 12, 2020"`.

:::

This text is outside of the fenced `div`.
```

You can modify [styles/common.css](styles/common.css) to apply styles to those fenced `div`s and their children:

```css
/* Style addenda. */
div.addendum {
  border: 1px solid grey;
  padding: 0 1em;
}

/* Include `date` attribute above addenda. */
div.addendum::before {
  display: block;
  text-align: center;
  color: grey;
  width: 100%;
  margin-top: 1em;
  content: "Addendum " attr(data-date);
}
```
