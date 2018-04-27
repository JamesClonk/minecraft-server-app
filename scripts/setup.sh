#!/bin/bash

# fail on error
set -e

# =============================================================================================
if [[ "$(basename $PWD)" == "scripts" ]]; then
	cd ..
fi
echo $PWD

# =============================================================================================
echo "downloading minecraft server ..."
wget https://launcher.mojang.com/mc/game/1.12.2/server/886945bfb2b978778c3a0288fd7fab09d315b25f/server.jar -O minecraft.jar

echo "downloading rcon cli ..."
wget https://github.com/itzg/rcon-cli/releases/download/1.3/rcon-cli_linux_amd64 -O rcon-cli
chmod +x rcon-cli
