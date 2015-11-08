#!/bin/env python3
import pathlib
import os
import re
import requests


def find_google_css(content):
    ret = []
    for i in re.finditer(r'@import url\("([\w\W]+?)"\);', content):
        ret.append([i.group(), i.groups()[0]])
    return ret


def new_file_name(url):
    return os.path.join("font", os.path.split(url)[1])


def fetch_google_css_and_font(font):
    origin, url = font
    print("deal with", url)
    headers = {"User-Agent": "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)"}
    # The google web font using woff for ie10, that's what we want
    content = requests.get(url, headers=headers).text
    for i in re.findall(r"url\(([\w\W]+?)\)", content):
        o = requests.get(i, stream=True)
        size = int(o.headers.get("Content-Length", 0))
        downlaoded = 0
        name = new_file_name(i)
        with open(name, "wb") as fp:
            print(i)
            for block in o.iter_content(1024):
                downlaoded += len(block)
                print("\rDownload {}/{}".format(downlaoded, size), end="")
                fp.write(block)
        print()
        content = content.replace(i, name)
    return content


def main():
    print("make sure you have good connection with google, or just set a HTTPS_PROXY env")
    p = pathlib.Path(".")
    files = list(p.glob("*.css"))
    if files == []:
        print("You should run the program in the vendor/themes folder")
        raise SystemExit
    os.makedirs("font", exist_ok=True)
    for i in files:
        print("Processing", i.name)
        content = i.read_text()
        csses = find_google_css(content)
        for css in csses:
            content = content.replace(css[0], fetch_google_css_and_font(css))
        print()
        i.write_text(content)

if __name__ == '__main__':
    main()
