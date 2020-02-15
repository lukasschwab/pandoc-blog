# pandoc-blog

## Requirements

+ `pandoc`
+ Python 3
+ Python dependencies listed in `requirements.txt`

## Usage

### Utilities

`make hook`: configure a git hook to run `make all` before each commit (so each commit contains an up-to-date static site).

`make requirements.txt`: install dependencies for `make_index.py`.
