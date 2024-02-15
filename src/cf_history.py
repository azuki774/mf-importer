import logging
import os
from pythonjsonlogger import jsonlogger
import mysql.connector
import format
import traceback

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
h = logging.StreamHandler()
h.setLevel(logging.DEBUG)
json_fmt = jsonlogger.JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s", json_ensure_ascii=False
)
h.setFormatter(json_fmt)
logger.addHandler(h)

def _dbClient():
    cnx = mysql.connector.connect(
        user="root",
        host=os.environ.get("db_host", "127.0.0.1"),
        database="mfimporter",
        password=os.getenv("db_pass"),
    )
    cnx.autocommit = False
    return cnx

def insert(insert_num, skip_num):
    logger.info("history record start")
    joblabel = os.environ.get("job_label", "")
    cnx = _dbClient()
    cur = cnx.cursor(buffered=True)

    try:
        insert_query = """
                       INSERT INTO import_history(
                            job_label,
                            parsed_entry_num,
                            new_entry_num
                       ) 
                       VALUES 
                       (%s, %s, %s);
                       """
        cur.execute(
            insert_query,
            (
                joblabel,
                insert_num + skip_num,
                insert_num,
            ),
        )
        cnx.commit()
    except Exception as e:
        traceback.print_exc()
        logger.info("failed to insert history: {0}".format(str(e)))
        cnx.rollback()
    finally:
        cnx.close()

    logger.info("history record end")
