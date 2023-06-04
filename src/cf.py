from bs4 import BeautifulSoup
import datetime

import logging
from pythonjsonlogger import jsonlogger
import pymongo
import queue

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
h = logging.StreamHandler()
h.setLevel(logging.DEBUG)
json_fmt = jsonlogger.JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s", json_ensure_ascii=False
)
h.setFormatter(json_fmt)
logger.addHandler(h)

FURIKAE_NAME = ["(カ) ジエ-シ-ビ-"]  # （振替）対象となり、フィールドの名称が異なるもの


def _dbClient():
    client = pymongo.MongoClient(
        "mf-importer-db",
        27017,
        username="root",
        password="pass",
    )
    return client


def get(filePath):
    date_num = 0
    with open(filePath) as f:
        html = f.read()
        today = datetime.date.today()  # 出力：datetime.date(2020, 3, 22)
        yyyymmdd = "{0:%Y-%m-%d}".format(today)  # 2020-03-22
        soup = BeautifulSoup(html, "html.parser")

    date_queue = queue.Queue()
    name_queue = queue.Queue()
    price_queue = queue.Queue()
    note_queue = queue.Queue()  # 保有金融機関
    v_l_ctg_queue = queue.Queue()  # 大分類
    v_m_ctg_queue = queue.Queue()  # 中分類

    tablea = soup.find_all(
        "td",
        {"class": "date"},
    )
    if len(tablea) >= 1:
        for e1 in tablea:
            if len(e1) >= 1:
                e2 = e1.findAll("span")
                for e3 in e2:
                    date_queue.put(e3.get_text().replace("\n", ""))
                    date_num = date_num + 1

    tableb = soup.find_all(
        "td",
        {"class": "content"},
    )
    if len(tableb) >= 1:
        for e1 in tableb:
            if len(e1) >= 1:
                e2 = e1.findAll("span")
                for e3 in e2:
                    name_queue.put(e3.get_text().replace("\n", ""))

    tablec = soup.find_all(
        "span",
        {"class": "offset"},
    )
    if len(tablec) >= 1:
        for e1 in tablec:
            if len(e1) >= 1:
                price_queue.put(e1.get_text().replace("\n", ""))

    tabled = soup.find_all(
        "td",
        {"class": "note calc"},  # （振替）のときは存在しないフィールド
    )
    if len(tabled) >= 1:
        for e1 in tabled:
            if len(e1) >= 1:
                note_queue.put(e1.get_text().replace("\n", ""))

    tablee = soup.find_all(
        "div",
        {"class": "btn-group btn_l_ctg"},  # （振替）のときは存在しないフィールド
    )
    if len(tablee) >= 1:
        for e1 in tablee:
            if len(e1) >= 1:
                v_l_ctg_queue.put(e1.get_text().replace("\n", ""))

    tablef = soup.find_all(
        "div",
        {"class": "btn-group btn_m_ctg"},  # （振替）のときは存在しないフィールド
    )
    if len(tablef) >= 1:
        for e1 in tablef:
            if len(e1) >= 1:
                v_m_ctg_queue.put(e1.get_text().replace("\n", ""))

    logger.info("parsed html file")
    inserted_data = []

    v_l_ctg_queue.get()  # 未分類がなぜか最初につくので空読みする
    v_m_ctg_queue.get()  # 未分類がなぜか最初につくので空読みする

    for i in range(date_num):
        ins_data = {}
        ins_data["regist_id"] = i + 1
        ins_data["regist_date"] = yyyymmdd
        ins_data["date"] = date_queue.get()
        ins_data["name"] = name_queue.get()
        ins_data["price"] = price_queue.get()

        # note calc フィールドなどは（振替）のときは存在しないのでパスする
        if ins_data["name"] not in FURIKAE_NAME:
            ins_data["fin_ins"] = note_queue.get()
            ins_data["l_category"] = v_l_ctg_queue.get()
            ins_data["m_category"] = v_m_ctg_queue.get()
        inserted_data.append(ins_data)

    if len(inserted_data) == 0:
        logger.info("No data in this file: {0}".format(filePath))
        return

    client = _dbClient()
    db = client.mfimporter
    collection_depo = db.detail
    collection_depo.insert_many(inserted_data)
    logger.info("inserted detail successfully")

    return inserted_data
