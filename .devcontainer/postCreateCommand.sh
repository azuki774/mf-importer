#!/bin/bash
WORKDIR="/workspace"

# DB migration
go install github.com/rubenv/sql-migrate/...@latest
MIGRATION_DIR="/workspace/mf-importer/migration"
cd ${MIGRATION_DIR}
sql-migrate up -env=dev
cd ${WORKDIR}

# useful symbolic link
sudo mkdir -p /data/
sudo chown -R "${USER}:${USER}" /data/
ln -s /workspace/mf-importer/.devcontainer/data/ /data

# python library
pip install --break-system-packages -r ${WORKDIR}/mf-importer/build/requirements.txt
