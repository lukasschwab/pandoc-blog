# pandoc-blog

Putting the "bodge" back in "blog:" Pandoc-driven static site generation.

## Requirements

+ `pandoc`
+ Python 3
+ Python dependencies listed in `requirements.txt`

## Usage

1. Clone this repository.
2. Write Pandoc-compatible Markdown files in `posts`. These should include YAML frontmatter for generating the index:
    + `title`: a human-readable title for this post.
    + `date`: some `dateutil`-parseable date format.
    + `abstract`: a summary you want to appear on the index. This can include valid Pandoc markdown.
3. Run `make all` to build an HTMl file for each Markdown page and generate `index.html`.

### Utilities

+ `make requirements.txt`: install dependencies for `make_index.py`.
+ `make hook`: configure a git hook to run `make all` before each commit (so each commit contains an up-to-date static site).

## How it works

`Makefile` is the most robust guide, but here's a high-level overview.

1. `pandoc` transforms each Markdown post in `posts` into a static HTML file in `gen`. The HTML is structured using `templates/post.html` and styled with `styles/shared.css`.

2. `make_index.py` reads the YAML frontmatter of every Markdown post in `posts` and transforms this into an intermediate Markdown document of headers and metadata, `index.md`.

3. `pandoc` transforms `index.md` into `index.html`. Unlike the posts, this index file is structured using `templates/index.html` and it's styled with *both* `styles/shared.css` and `styles/index.css` (with the latter styles overriding the former.

## Customization

A general rule of thumb: changes to the HTML are predictable; changes to pre-`pandoc` Markdown are unpredictable. Markdown intermediates (like `make_index.py` uses for the time being) are antipatterns.

+ Want to change how posts are represented in the index?<br>Modify `make_index.py`.

+ Want to add static elements, e.g. a section with "about me" info or social links?<br>Modify `templates/index.html` to only change the index.<br>Modify `templates/post.html` to only change the post pages.

+ Want to change how the index is styled?<br>Modify `styles/index.css`.

+ Want to change how the whole generated site is styled?<br>Modify `styles/common.css`.

## To do

`make_index.py` is super brittle. It should be extended to read a greater variety of pandoc-supported YAML frontmatter and read full-blog metadata defined in some root YAML file.
