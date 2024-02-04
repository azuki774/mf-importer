import datetime
from argparse import ArgumentParser
import logging
from pythonjsonlogger import jsonlogger
import cf

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
h = logging.StreamHandler()
h.setLevel(logging.DEBUG)
json_fmt = jsonlogger.JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s", json_ensure_ascii=False
)
h.setFormatter(json_fmt)
logger.addHandler(h)

DATA_BASE_DIR = "/data/"


def main():
    """
    htmlページをパースして、DBに情報を格納する。
    引数に対象のページ名を入れる。
    main.py <cf|portofolio>
    読み込みディレクトリは、/data/<file名>
    """

    logger.info("start")
    argparser = ArgumentParser()
    argparser.add_argument(
        "component",
        type=str,
        help="also fetch lastmonth data",
    )
    argparser.add_argument(
        "--lastmonth",
        type=bool,
        default=False,
        help="also fetch lastmonth data",
    )
    argp = argparser.parse_args()


    if argp.component == "cf":
        # --last-month があれば cf_lastmonth も
        ret = cf.get(DATA_BASE_DIR + "/cf")
        if ret != 0:
            return ret  # error end
        if argp.lastmonth:
            logger.info("lastmonth option detected")
            ret = cf.get(DATA_BASE_DIR + "/cf_lastmonth")
            if ret != 0:
                return ret  # error end

    logger.info("end")


if __name__ == "__main__":
    main()
