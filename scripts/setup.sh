#!/bin/bash

# fail on error
set -e

# =============================================================================================
if [[ "$(basename $PWD)" == "scripts" ]]; then
	cd ..
fi
echo $PWD

# =============================================================================================
echo "downloading minecraft server (1.14.4) ..."
wget https://launcher.mojang.com/v1/objects/3dc3d84a581f14691199cf6831b71ed1296a9fdf/server.jar -O minecraft.jar

echo "downloading rcon cli ..."
wget https://github.com/itzg/rcon-cli/releases/download/1.4.6/rcon-cli_1.4.6_linux_amd64.tar.gz -O rcon-cli.tar.gz
tar -xvzf rcon-cli.tar.gz
chmod +x rcon-cli

echo "downloading minio client ..."
wget https://dl.minio.io/client/mc/release/linux-amd64/mc -O mc
chmod +x mc
