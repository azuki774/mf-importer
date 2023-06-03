FROM python:3.8-slim-buster as runner
RUN pip3 install bs4 python-json-logger pymongo
WORKDIR /src
COPY src/ /src
ENTRYPOINT ["python3", "main.py"]
