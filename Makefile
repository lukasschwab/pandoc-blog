POSTS=$(shell find posts/*)
# OUT contains all names of static HTML targets corresponding to markdown files
# in the posts directory.
OUT=$(patsubst posts/%.md, static/%.html, $(POSTS))

all: $(OUT) index.html

# TODO: template.
static/%.html: posts/%.md
	pandoc -s $< -o $@

# TODO: template.
index.html: $(OUT)
	# Building index.html.
	python3 make_index.py
	pandoc -s index.html -o index.html --metadata pagetitle="blog"

clean:
	rm -f static/*.html

install: requirements.txt
	ln -s -f ../../hooks/pre-commit ./.git/hooks/pre-commit

requirements.txt:
	pip install -r requirements.txt
