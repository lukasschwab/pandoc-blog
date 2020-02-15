# pandoc-blog

Putting the "bodge" back in "blog:" Pandoc-driven static site generation.

## Requirements

+ `pandoc`
+ Python 3
+ Python dependencies listed in `requirements.txt`

## Usage

1. Write Pandoc-compatible Markdown files in `posts`.
2. Run `make all` to build an HTMl file for each Markdown page and generate `index.html`.

### Utilities

+ `make requirements.txt`: install dependencies for `make_index.py`.
+ `make hook`: configure a git hook to run `make all` before each commit (so each commit contains an up-to-date static site).

## Notes

Instead of generating an index with a script, I could use a manually-maintained index with a definition list:

```markdown
# blog

Here's a definition list
: And here's the summary for the definition list. Who knows if it'll work?

[Here's a link](./gen/some-post.html)
: And here's a definition for the link.
```
