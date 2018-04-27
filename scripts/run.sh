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
source .env_private

# =============================================================================================
echo "running minecraft-server-app ..."
java -jar launcher.jar minecraft-server-app
