from bs4 import BeautifulSoup
import datetime

import logging
import os
from pythonjsonlogger import jsonlogger
import pymongo
import format

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
h = logging.StreamHandler()
h.setLevel(logging.DEBUG)
json_fmt = jsonlogger.JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s", json_ensure_ascii=False
)
h.setFormatter(json_fmt)
logger.addHandler(h)

FURIKAE_NAME = [
    "(カ) ジエ-シ-ビ-",
    "ミツイスミトモカ-ド (カ",
    "SBI証券投信積立サービス",
]  # （振替）対象となり、フィールドの名称が異なるもの


def _dbClient():
    client = pymongo.MongoClient(
        "mf-importer-db",
        27017,
        username="root",
        password=os.getenv("db_pass"),
    )
    return client


def get(filePath):
    date_num = 0
    with open(filePath) as f:
        html = f.read()
        today = datetime.date.today()  # 出力：datetime.date(2020, 3, 22)
        proc_yyyymmdd = "{0:%Y%m%d}".format(today)  # 処理日のYYYYMMDD 20200322
        soup = BeautifulSoup(html, "html.parser")

    date_list = []
    name_list = []
    price_list = []
    note_list = []  # 保有金融機関
    v_l_ctg_list = []  # 大分類
    v_m_ctg_list = []  # 中分類

    tablea = soup.find_all(
        "td",
        {"class": "date"},
    )
    if len(tablea) >= 1:
        for e1 in tablea:
            if len(e1) >= 1:
                e2 = e1.findAll("span")
                for e3 in e2:
                    date_list.append(e3.get_text().replace("\n", ""))
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
                    name_list.append(e3.get_text().replace("\n", ""))

    tablec = soup.find_all(
        "span",
        {"class": "offset"},
    )
    if len(tablec) >= 1:
        for e1 in tablec:
            if len(e1) >= 1:
                price_list.append(e1.get_text().replace("\n", ""))

    tabled = soup.find_all(
        "td",
        {"class": "note calc"},  # （振替）のときは存在しないフィールド
    )
    if len(tabled) >= 1:
        for e1 in tabled:
            if len(e1) >= 1:
                note_list.append(e1.get_text().replace("\n", ""))

    tablee = soup.find_all(
        "div",
        {"class": "btn-group btn_l_ctg"},  # （振替）のときは存在しないフィールド
    )
    if len(tablee) >= 1:
        for e1 in tablee:
            if len(e1) >= 1:
                v_l_ctg_list.append(e1.get_text().replace("\n", ""))

    tablef = soup.find_all(
        "div",
        {"class": "btn-group btn_m_ctg"},  # （振替）のときは存在しないフィールド
    )
    if len(tablef) >= 1:
        for e1 in tablef:
            if len(e1) >= 1:
                v_m_ctg_list.append(e1.get_text().replace("\n", ""))

    logger.info("parsed html file")
    inserted_data = []

    for i in range(date_num):
        ins_data = {}
        ins_data["yyyymm_id"] = i + 1  # 利用月ごとのID
        ins_data["raw_date"] = date_list.pop()  # 日付昇順で regist_id を振るため、pop を利用する
        ins_data["date"] = format.get_yyyymmdd_from_procdate(
            proc_yyyymmdd, ins_data["raw_date"]
        )
        ins_data["name"] = name_list.pop()
        ins_data["raw_price"] = price_list.pop()
        ins_data["price"] = format.get_price_frow_raw(ins_data["raw_price"])

        # note calc フィールドなどは（振替）のときは存在しないのでパスする
        if ins_data["name"] not in FURIKAE_NAME:
            ins_data["fin_ins"] = note_list.pop()
            ins_data["l_category"] = v_l_ctg_list.pop()
            ins_data["m_category"] = v_m_ctg_list.pop()

        ins_data["regist_date"] = proc_yyyymmdd
        inserted_data.append(ins_data)

    v_l_ctg_list.pop()  # 未分類がなぜか最初につくので、最後に pop にする
    v_m_ctg_list.pop()  # 未分類がなぜか最初につくので、最後に pop にする

    if len(inserted_data) == 0:
        logger.info("No data in this file: {0}".format(filePath))
        return 0

    if len(v_l_ctg_list) != 0:
        logger.error("failed to parse")
        return 1

    if len(v_m_ctg_list) != 0:
        logger.error("failed to parse")
        return 1

    _insert(inserted_data)
    return 0


def _insert(insert_data):
    client = _dbClient()
    db = client.mfimporter
    collection_depo = db.detail

    loaded_num = 0
    insert_num = 0
    for data in insert_data:
        find = collection_depo.find_one(
            # yyyymmdd と name と price がすべて一致するものを抽出する（登録済と判断する）
            filter={
                "raw_date": data["raw_date"],
                "name": data["name"],
                "raw_price": data["raw_price"],
            }
        )
        if find == None:
            # 未登録なら 登録
            collection_depo.insert_one(data)
            insert_num += 1

        loaded_num += 1

    logger.info("inserted new detail successfully: {0} records".format(insert_num))
    logger.info(
        "skipped already inserted records: {0} records".format(loaded_num - insert_num)
    )
    return 0
