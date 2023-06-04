import sys
from argparse import ArgumentParser
import logging
from pythonjsonlogger import jsonlogger
import portfolio
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


def main():
    """
    htmlページをパースして、DBに情報を格納する。
    引数に対象のページ名を入れる。
    main.py <cf|portofolio>
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

    # if len(args) <= 1:
    #     logger.error("required args")
    #     return

    if argp.component == "portfolio":
        portfolio.get("/data/portfolio")

    if argp.component == "cf":
        # --last-month があれば cf_lastmonth も
        cf.get("/data/cf")
        if argp.lastmonth:
            logger.info("lastmonth option detected")
            cf.get("/data/cf_lastmonth")

    logger.info("end")


if __name__ == "__main__":
    main()
