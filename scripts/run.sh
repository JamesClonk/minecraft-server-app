#!/bin/bash

# fail on error
set -e

# =============================================================================================
if [[ "$(basename $PWD)" == "scripts" ]]; then
	cd ..
fi
echo $PWD

# =============================================================================================
source .env

# =============================================================================================
echo "downloading minecraft server ..."
wget https://launcher.mojang.com/mc/game/1.12.2/server/886945bfb2b978778c3a0288fd7fab09d315b25f/server.jar -O minecraft.jar

# =============================================================================================
echo "running minecraft-server-app ..."
./minecraft-server-app
