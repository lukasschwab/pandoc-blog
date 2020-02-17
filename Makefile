POSTS=$(shell find posts/*)
# OUT contains all names of static HTML targets corresponding to markdown files
# in the posts directory.
OUT=$(patsubst posts/%.md, gen/%.html, $(POSTS))

all: $(OUT) index.html

gen/%.html: posts/%.md
	pandoc -s $< -o $@ --template templates/post.html --css="../styles/common.css"

index.html: $(OUT)
	python3 make_index.py
	pandoc -s index.md -o index.html --template templates/index.html  --css="./styles/common.css" --css="./styles/index.css"
	rm index.md

# Shortcuts

open: all
	open index.html

# Get an ISO 8601 date.
date:
	date -u +"%Y-%m-%dT%H:%M:%SZ"

clean:
	rm -f gen/*.html

hook:
	ln -s -f ../../.hooks/pre-commit ./.git/hooks/pre-commit

.PHONY: install
install:
	pip install -r requirements.txt
