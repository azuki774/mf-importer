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


def _dbClient():
    client = pymongo.MongoClient(
        "mf-importer-db",
        27017,
        username="root",
        password="pass",
    )
    return client


def get(filePath):
    with open(filePath) as f:
        html = f.read()
        today = datetime.date.today()  # 出力：datetime.date(2020, 3, 22)
        yyyymmdd = "{0:%Y-%m-%d}".format(today)  # 2020-03-22
        soup = BeautifulSoup(html, "html.parser")

    depo_dicts = []
    mf_dicts = []
    pns_dicts = []

    # 預金・現金・暗号資産
    table = soup.find_all("table", {"class": "table table-bordered table-depo"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 5):
        d_dict = {}
        d_dict["regist_id"] = i + 1
        d_dict["regist_date"] = yyyymmdd
        d_dict["category"] = "預金・現金・暗号資産"
        d_dict["種類・名称"] = elements[5 * i + 0].get_text()
        d_dict["残高"] = elements[5 * i + 1].get_text()
        d_dict["保有金融機関"] = elements[5 * i + 2].get_text()
        depo_dicts.append(d_dict)
    logger.info("loaded table-depo")

    # 投資信託
    table = soup.find_all("table", {"class": "table table-bordered table-mf"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 12):
        mf_dict = {}
        mf_dict["regist_id"] = i + 1
        mf_dict["regist_date"] = yyyymmdd
        mf_dict["category"] = "投資信託"
        mf_dict["銘柄名"] = elements[12 * i + 0].get_text().replace("\n", "")
        mf_dict["保有数"] = elements[12 * i + 1].get_text().replace("\n", "")
        mf_dict["平均取得単価"] = elements[12 * i + 2].get_text().replace("\n", "")
        mf_dict["基準価額"] = elements[12 * i + 3].get_text().replace("\n", "")
        mf_dict["評価額"] = elements[12 * i + 4].get_text().replace("\n", "")
        mf_dict["前日比"] = elements[12 * i + 5].get_text().replace("\n", "")
        mf_dict["評価損益"] = elements[12 * i + 6].get_text().replace("\n", "")
        mf_dict["保有金融機関"] = elements[12 * i + 7].get_text().replace("\n", "")
        mf_dict["取得日"] = elements[12 * i + 8].get_text().replace("\n", "")
        # print(elements[12 * i + 9].get_text()) # blank
        mf_dicts.append(mf_dict)
    logger.info("loaded table-mf")

    # ポイント・マイル
    table = soup.find_all("table", {"class": "table table-bordered table-pns"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 8):
        pns_dict = {}
        pns_dict["regist_id"] = i + 1
        pns_dict["regist_date"] = yyyymmdd
        pns_dict["category"] = "ポイント・マイル"
        pns_dict["名称"] = elements[8 * i + 0].get_text().replace("\n", "")
        pns_dict["種類"] = elements[8 * i + 1].get_text().replace("\n", "")
        pns_dict["ポイント・マイル数"] = elements[8 * i + 2].get_text().replace("\n", "")
        pns_dict["換算レート"] = elements[8 * i + 3].get_text().replace("\n", "")
        pns_dict["現在の価値"] = elements[8 * i + 4].get_text().replace("\n", "")
        pns_dict["保有金融機関"] = elements[8 * i + 5].get_text().replace("\n", "")
        pns_dicts.append(pns_dict)
    logger.info("loaded table-pns")

    print(depo_dicts)
    print(mf_dicts)
    print(pns_dicts)

    client = _dbClient()

    db = client.mfimporter
    collection_depo = db.depo
    collection_depo.insert_many(depo_dicts)
    logger.info("inserted table-depo successfully")

    collection_mf = db.mf
    collection_mf.insert_many(mf_dicts)
    logger.info("inserted table-mf successfully")

    collection_pns = db.pns
    collection_pns.insert_many(pns_dicts)
    logger.info("inserted table-pns successfully")
