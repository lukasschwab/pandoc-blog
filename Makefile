all: build
	./pandoc-blog

build:
	go build -o pandoc-blog .

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

.PHONY: all build open date clean hook
