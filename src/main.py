import sys
import portfolio
import logging
from pythonjsonlogger import jsonlogger

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

    args = sys.argv
    logger.info("start")

    if len(args) <= 1:
        logger.error("required args")

    if args[1] == "portfolio":
        portfolio.get("/data/portfolio")

    logger.info("end")


if __name__ == "__main__":
    main()
