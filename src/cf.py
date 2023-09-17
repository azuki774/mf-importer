from bs4 import BeautifulSoup
import datetime

import logging
import os
from pythonjsonlogger import jsonlogger
import mysql.connector
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
    cnx = mysql.connector.connect(
        user="root", database="mfimporter", password=os.getenv("pass")
    )
    cnx.autocommit = False
    return cnx


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
    """
    引数の insert_data を実際にDB - detailに挿入する。
    その仮定で既に登録済だと思われるものは省略する。
    [insert_num, skip_num] を返す。
    """
    cnx = _dbClient()
    cur = cnx.cursor(buffered=True)

    loaded_num = 0
    insert_num = 0

    try:
        for data in insert_data:
            loaded_num += 1
            # yyyymmdd と name と price がすべて一致するものを抽出する（登録済と判断するため）
            pre_search_query = (
                "SELECT count(1) FROM detail "
                "WHERE raw_date = %s "
                "AND   name = %s "
                "AND   raw_price = %s "
            )

            cur.execute(
                pre_search_query, (data["raw_date"], data["name"], data["raw_price"])
            )
            rows = cur.fetchall()
            num = rows[0][0]  # count(1) の結果の数字を取得
            if num == 0:
                # 未登録なら 登録
                insert_num += 1
        cnx.commit()
    except Exception as e:
        logger.info("failed to insert records: {0}".format(str(e)))
        cnx.rollback()
    finally:
        cnx.close()

    skip_num = loaded_num - insert_num

    logger.info("inserted new detail successfully: {0} records".format(insert_num))
    logger.info("skipped already inserted records: {0} records".format(skip_num))
    return [insert_num, skip_num]
