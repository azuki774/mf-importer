from bs4 import BeautifulSoup
import datetime

import logging
from pythonjsonlogger import jsonlogger
import pymongo

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
h = logging.StreamHandler()
h.setLevel(logging.DEBUG)
json_fmt = jsonlogger.JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s", json_ensure_ascii=False
)
h.setFormatter(json_fmt)
logger.addHandler(h)


def get(filePath):
    counta = 0
    countb = 0
    countc = 0
    countd = 0
    with open(filePath) as f:
        html = f.read()
        today = datetime.date.today()  # 出力：datetime.date(2020, 3, 22)
        yyyymmdd = "{0:%Y-%m-%d}".format(today)  # 2020-03-22
        soup = BeautifulSoup(html, "html.parser")
    tablea = soup.find_all(
        "table",
        {"class": "table table-hover table-autosort:1 table-autosort table-autofilter"},
    )
    if len(tablea) >= 1:
        table = tablea[0]
        elements = table.findAll("span")
        for e in elements:
            print(e.get_text().replace("\n", ""))
            counta = counta + 1

    tableb = soup.find_all("td", {"class": "note calc"})
    print("---------------------------------------------")
    for e in tableb:
        print(e.get_text().replace("\n", ""))  # 保有金融機関
        countb = countb + 1

    print("---------------------------------------------")
    tablec = soup.find_all("div", {"class": "btn-group btn_l_ctg"})
    for e in tablec:
        print(e.get_text().replace("\n", ""))  # 大分類
        countc = countc + 1

    print("---------------------------------------------")
    tablec = soup.find_all("div", {"class": "btn-group btn_m_ctg"})
    for e in tablec:
        print(e.get_text().replace("\n", ""))  # 中分類
        countd = countd + 1

    print(counta)
    print(countb)
    print(countc)
    print(countd)
