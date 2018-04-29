#!/bin/bash

# fail on error
set -e

# =============================================================================================
if [[ "$(basename $PWD)" == "scripts" ]]; then
	cd ..
fi
echo $PWD

# =============================================================================================
echo "backing up minecraft world ..."
./rcon-cli --port 25575 --password ${MINECRAFT_RCON_PASSWORD} save-off
./rcon-cli --port 25575 --password ${MINECRAFT_RCON_PASSWORD} save-all
sleep 5
tar -cvzf world-backup.tar.gz world/
sleep 1
./rcon-cli --port 25575 --password ${MINECRAFT_RCON_PASSWORD} save-on

# =============================================================================================
echo "uploading minecraft world backup ..."
./mc cp world-backup.tar.gz s3/${MINECRAFT_BACKUP_BUCKET_NAME}/world-backup.tar.gz
