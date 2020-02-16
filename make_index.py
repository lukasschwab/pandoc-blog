# make_index.py generates a Markdown document representing the blog index.
# Ideally, this should only contain the list of links to posts and whatever
# metadata should be listed in those posts. For header/footer information, it
# will probably be easier to modify templates/index.html.

from os import listdir
import frontmatter
from datetime import datetime
from dateutil.parser import parse
import pytz
import jsonfeed as jf

MARKDOWN_POSTS_PATH = "./posts"
STATIC_FILES_PATH = "./gen/"

isNaive = lambda d: d.tzinfo is None or d.tzinfo.utcoffset(d) is None

def getMetadata(static_post):
    post = frontmatter.load(MARKDOWN_POSTS_PATH + "/" + static_post)
    out = {}
    out['title'] = post.get('title', static_post[:-3])

    if 'date' in post:
        out['date'] = parse(post.get('date'))
    else:
        print("[WARN] no date for post", static_post)
        out['date'] = datetime.now(tz=pytz.UTC)
    # If we don't have timezone info, we assume UTC.
    if isNaive(out['date']):
        print("[WARN] No time zone for date; assuming UTC")
        out['date'] = out['date'].replace(tzinfo=pytz.UTC)

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

# Construct a JSON feed.
feed = jf.Feed("blog")
for metadata in metadatas:
    url = getStaticFilename(metadata)
    item = jf.Item(
        url, # url as primary ID.
        url=url,
        title=metadata['title']
    )
    # These components may or may not be defined.
    if metadata['date']:
        # isoformat is RFC-compatible iff the datetime is timezone-aware.
        item.date_published = metadata['date'].isoformat()
    # Junky default
    item.content_text = metadata['title']
    if metadata['abstract']:
        item.summary = metadata['abstract']
        item.content_text = metadata['abstract']

    feed.items.append(item)

with open("./feed.json", "w") as f:
    f.write(feed.toJSON())
