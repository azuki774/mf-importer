import cf


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
        assert t["want"] == cf._get_yyyymmdd_from_procdate(
            t["proc_yyyymmdd"], t["date"]
        )
