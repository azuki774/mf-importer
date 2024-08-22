FROM node:22

RUN apt-get -y update

RUN mkdir /data
RUN npm i -g npm
RUN npm i -g redoc-cli

ENTRYPOINT ["redoc-cli", "bundle", "/data/mfimporter-api.yaml", "--output", "/data/api.html"]
