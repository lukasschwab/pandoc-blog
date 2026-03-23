POSTS=$(shell find posts/*)
# OUT contains all names of static HTML targets corresponding to markdown files
# in the posts directory.
OUT=$(patsubst posts/%.md, gen/%.html, $(POSTS))

all: $(OUT) index.html

gen/%.html: posts/%.md
	pandoc -f markdown+fenced_divs -s $< -o $@ --template templates/post.html --css="../styles/common.css"

pandoc-blog: main.go go.mod go.sum
	go build -o pandoc-blog .

index.html: $(OUT) pandoc-blog
	./pandoc-blog

# Shortcuts

open: all
	open index.html

# Get an ISO 8601 date.
date:
	date -u +"%Y-%m-%dT%H:%M:%SZ"

clean:
	rm -f gen/*.html
	rm -f index.html
	rm -f feed.json
	rm -f pandoc-blog

hook:
	ln -s -f ../../.hooks/pre-commit ./.git/hooks/pre-commit

.PHONY: open date clean hook
