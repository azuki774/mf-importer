import format


def test_get_yyyymmdd_from_procdate():
    tt = [
        {
            "proc_yyyymmdd": "20230519",
            "date": "05/16（火）",
            "want": "20230516",
        },
        {
            "proc_yyyymmdd": "20230102",
            "date": "12/01（火）",
            "want": "20221201",
        },
    ]

    for t in tt:
        assert t["want"] == format.get_yyyymmdd_from_procdate(
            t["proc_yyyymmdd"], t["date"]
        )


def test_get_price_frow_raw():
    tt = [
        {
            "raw_price": "1,230",
            "want": 1230,
        },
        {
            "raw_price": "-23,450",
            "want": 23450,
        },
        {
            "raw_price": "-45",
            "want": 45,
        },
    ]

    for t in tt:
        assert t["want"] == format.get_price_frow_raw(t["raw_price"])
