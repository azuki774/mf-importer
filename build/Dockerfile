FROM python:3.8-slim-buster as runner
RUN pip3 install bs4
WORKDIR /src
COPY src/ /src
ENTRYPOINT ["python3"]
CMD ["main.py"]
