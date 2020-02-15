POSTS=$(shell find posts/*)
# OUT contains all names of static HTML targets corresponding to markdown files
# in the posts directory.
OUT=$(patsubst posts/%.md, gen/%.html, $(POSTS))

all: $(OUT) index.html

# TODO: template.
gen/%.html: posts/%.md
	pandoc -s $< -o $@

# TODO: template.
index.html: $(OUT)
	# Building index.html.
	python3 make_index.py
	pandoc -s index.html -o index.html --metadata pagetitle="blog"

clean:
	rm -f gen/*.html

hook:
	ln -s -f ../../.hooks/pre-commit ./.git/hooks/pre-commit

requirements.txt:
	pip install -r requirements.txt
