from bs4 import BeautifulSoup


def get(filePath):
    with open(filePath) as f:
        html = f.read()
        # today = datetime.date.today()  # 出力：datetime.date(2020, 3, 22)
        # yyyymmdd = "{0:%Y%m%d}".format(today)  # 20200322
        soup = BeautifulSoup(html, "html.parser")

    # 預金・現金・暗号資産
    table = soup.find_all("table", {"class": "table table-bordered table-depo"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 5):
        print(elements[5 * i + 0].get_text())  # 種類・名称
        print(elements[5 * i + 1].get_text())  # 残高
        print(elements[5 * i + 2].get_text())  # 保有金融機関

    # 投資信託
    table = soup.find_all("table", {"class": "table table-bordered table-mf"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 12):
        print(elements[12 * i + 0].get_text())
        print(elements[12 * i + 1].get_text())
        print(elements[12 * i + 2].get_text())
        print(elements[12 * i + 3].get_text())
        print(elements[12 * i + 4].get_text())
        print(elements[12 * i + 5].get_text())
        print(elements[12 * i + 6].get_text())
        print(elements[12 * i + 7].get_text())
        print(elements[12 * i + 8].get_text())
        # print(elements[12 * i + 9].get_text()) # blank

    # ポイント・マイル
    table = soup.find_all("table", {"class": "table table-bordered table-pns"})[0]
    elements = table.findAll("td")
    for i in range(len(elements) // 8):
        print(elements[8 * i + 0].get_text())
        print(elements[8 * i + 1].get_text())
        print(elements[8 * i + 2].get_text())
        print(elements[8 * i + 3].get_text())
        print(elements[8 * i + 4].get_text())
        print(elements[8 * i + 5].get_text())
