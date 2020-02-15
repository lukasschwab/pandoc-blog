POSTS=$(shell find posts/*)
# OUT contains all names of static HTML targets corresponding to markdown files
# in the posts directory.
OUT=$(patsubst posts/%.md, gen/%.html, $(POSTS))

all: $(OUT) index.html

# TODO: template.
gen/%.html: posts/%.md
	pandoc -s $< -o $@ --template templates/post.html --css="../common.css"

# TODO: template.
index.html: $(OUT)
	# Building index.html.
	python3 make_index.py
	pandoc -s index.html -o index.html --template templates/index.html  --css="./common.css"

open: all
	open index.html

clean:
	rm -f gen/*.html

hook:
	ln -s -f ../../.hooks/pre-commit ./.git/hooks/pre-commit

requirements.txt:
	pip install -r requirements.txt
