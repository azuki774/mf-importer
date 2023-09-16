def get_yyyymmdd_from_procdate(proc_yyyymmdd, date):
    """
    proc_yyyymmdd をベースにして、date フィールドから yyyymmdd を取得する
    date フィールドは '05/09(火)' のような表記のため yyyy 情報が必要

    proc_yyyymmdd = 20230519 , 05/15（ ）なら 20230515
    proc_yyyymmdd = 20230101 , 12/15（ ）なら 20221215
    """

    proc_yyyymmdd_yyyy = proc_yyyymmdd[:4]
    proc_yyyymmdd_mm = proc_yyyymmdd[4:6]
    date_mm = date[:2]
    date_dd = date[3:5]

    if proc_yyyymmdd_mm == "01" and date_mm == "12":
        # 年が変わるパターンだけ例外
        return str(int(proc_yyyymmdd_yyyy) - 1) + date_mm + date_dd

    # 通常パターン
    return proc_yyyymmdd_yyyy + date_mm + date_dd


def get_price_frow_raw(raw_price):
    """
    '-1,250'みたいなやつを1250にする
    """

    price = raw_price.replace("-", "")
    price = price.replace(",", "")
    return int(price)
