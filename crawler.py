import time
import asyncio
import aiohttp
import argparse
from bs4 import BeautifulSoup, CData


def chunks(lst, n):
    """Yield successive n-sized chunks from lst."""
    for i in range(0, len(lst), n):
        yield lst[i:i + n]


def parse_xml_file(filename):
    xml_data = None
    try:
        with open(filename, 'r') as xml_file:
            xml_data = xml_file.read()
    except FileNotFoundError:
        return None

    items = []
    soup = BeautifulSoup(xml_data, "html.parser")
    for url in soup.findAll("url"):
        temp = {
            "url": url.find("loc").text,
            "title": url.find("news:title"),
            "publication_date": url.find("news:publication_date").text,
            "keywords": url.find("news:keywords")
        }
        for cd in temp["title"].findAll(text=True):
            if isinstance(cd, CData):
                temp["title"] = cd

        for cd in temp["keywords"].findAll(text=True):
            if isinstance(cd, CData):
                temp["keywords"] = cd
        items.append(temp)
    return items


async def fetch(session, url):
    async with session.get(url) as resp:
        print("[crawler] {} : {}".format(url, resp.status))
        return {
            "url": url,
            "status_code": resp.status
        }


async def fetch_all(urls, loop):
    async with aiohttp.ClientSession(loop=loop) as session:
        results = await asyncio.gather(*[fetch(session, url) for \
            url in urls], return_exceptions=True)
        return results


def crawl(items):
    loop = asyncio.get_event_loop()
    total_data = len(items)
    success = 0
    for chunked_items in chunks(items, 10):
        urls = [i['url'] for i in chunked_items]
        results = loop.run_until_complete(fetch_all(urls, loop))
        for result in results:
            if result['status_code'] == 200:
                success += 1
        time.sleep(0.5)
    print("total data {}, success {}".format(total_data, success))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--filename", type=str)
    args = parser.parse_args()

    filename = args.filename
    items = parse_xml_file(filename)

    start_time = time.time()
    crawl(items)
    elapsed = time.time() - start_time
    print("crawl took {} second".format(elapsed))