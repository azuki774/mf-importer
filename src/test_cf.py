import cf
import os


def test_insert():
    # prepare testdata
    os.environ["pass"] = "password"
    cnx = cf._dbClient()
    cur = cnx.cursor(buffered=True)

    ## 必要なものだけ入れる
    queries = [
        'INSERT INTO detail(id, yyyymm_id, date, name, raw_date, raw_price, regist_date) VALUES (100, 1, "2010-01-01", "テスト", "01/01(火)", "-1,234", "2010-01-04");',
        'INSERT INTO detail(id, yyyymm_id, date, name, raw_date, raw_price, regist_date) VALUES (101, 2, "2010-01-02", "テスト", "01/02(水)", "-123", "2010-01-04");',
    ]
    for q in queries:
        cur.execute(q)

    cnx.commit()
    cnx.close()

    tt = [
        {  # Case 1: id = 1,2は登録済で 3を新規登録
            "insert_data": [
                {
                    "yyyymm_id": 1,
                    "raw_date": "01/01(火)",
                    "date": "2010-01-01",
                    "name": "テスト",
                    "raw_price": "-1,234",
                    "price": 1234,
                    "fin_ins": "",
                    "l_category": "",
                    "m_category": "",
                    "regist_date": "2010-01-04",
                },
                {
                    "yyyymm_id": 2,
                    "raw_date": "01/02(水)",
                    "date": "2010-01-02",
                    "name": "テスト",
                    "raw_price": "-123",
                    "price": 123,
                    "fin_ins": "",
                    "l_category": "",
                    "m_category": "",
                    "regist_date": "2010-01-04",
                },
                {
                    "yyyymm_id": 3,
                    "raw_date": "01/02(水)",
                    "date": "2010-01-02",
                    "name": "テストINSERT",
                    "raw_price": "-123,456",
                    "price": 123456,
                    "fin_ins": "",
                    "l_category": "",
                    "m_category": "",
                    "regist_date": "2010-01-05",
                },
            ],
            "want": [1, 2],
        }
    ]

    try:
        for t in tt:
            assert t["want"] == cf.insert(t["insert_data"])
    finally:
        # delete testdata
        os.environ["pass"] = "password"
        cnx = cf._dbClient()
        cur = cnx.cursor(buffered=True)

        queries = [
            "DELETE FROM detail WHERE id = 100 OR id = 101;",
            "DELETE FROM detail WHERE name = 'テストINSERT'",
        ]
        for q in queries:
            cur.execute(q)

        cnx.commit()
        cnx.close()
