from os import listdir

STATIC_FILES_PATH = "./static"

def linkToPost(static_post):
    return '<a href="static/{}">{}</a>'.format(static_post, static_post)

# TODO: sort by Markdown metadata date.
# TODO: description list with summaries. Can use the abstract YAML attr.
if __name__ == "__main__":
    static_posts = listdir(STATIC_FILES_PATH)
    with open("./index.html", "w") as f:
        f.write("<ol>\n")
        for static_post in static_posts:
            f.write("<li>{}</li>\n".format(linkToPost(static_post)))
        f.write("</ol>\n")
