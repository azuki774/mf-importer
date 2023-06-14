FROM python:3.8-slim-buster as runner
WORKDIR /src
COPY src/ /src
COPY build/requirements.txt /src/requirements.txt
RUN pip install -r /src/requirements.txt
ENTRYPOINT ["python3", "main.py"]
