# make_index.py generates a Markdown document representing the blog index.
# Ideally, this should only contain the list of links to posts and whatever
# metadata should be listed in those posts. For header/footer information, it
# will probably be easier to modify templates/index.html.

from os import listdir
import frontmatter
from datetime import datetime
from dateutil.parser import parse

MARKDOWN_POSTS_PATH = "./posts"
STATIC_FILES_PATH = "./gen/"

def getMetadata(static_post):
    post = frontmatter.load(MARKDOWN_POSTS_PATH + "/" + static_post)
    out = {}
    out['title'] = post.get('title', static_post[:-3])
    # NOTE: stripping time zone info is a bodge.
    if 'date' in post:
        out['date'] = parse(post.get('date')).replace(tzinfo=None)
    else:
        print("[WARN] no date for post", static_post)
        out['date'] = datetime.now().replace(tzinfo=None)
    out['abstract'] = post.get('abstract')
    out['filename'] = static_post
    return out

def getStaticFilename(post_metadata):
    return STATIC_FILES_PATH + post_metadata['filename'][:-3] + ".html"

static_posts = listdir(MARKDOWN_POSTS_PATH)
metadatas = [getMetadata(fn) for fn in static_posts]
metadatas.sort(key=lambda md: md['date'], reverse=True)

with open("./index.md", "w") as f:
    if len(metadatas) == 0:
        f.write("There aren't any posts yet.\n")
    else:
        for metadata in metadatas:
            f.write("## [{}]({})\n".format(
                metadata['title'],
                getStaticFilename(metadata)
            ))
            if metadata['abstract']:
                f.write("{} &middot; {}\n".format(
                    metadata['date'].strftime("%Y-%m-%d"),
                    metadata['abstract']
                ))
            f.write("\n")
