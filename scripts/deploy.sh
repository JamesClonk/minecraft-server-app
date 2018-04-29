#!/bin/bash

# fail on error
set -e

# =============================================================================================
if [[ "$(basename $PWD)" == "scripts" ]]; then
	cd ..
fi
echo $PWD

# =============================================================================================
echo "deploying minecraft-server-app ..."
cf push

# # =============================================================================================
# echo "setting up routes and ports ..."
# cf create-route PROD tcp.ci1-ares-031.swisslab.io --port 31337
# export TCP_ROUTE_GUID=$(CF_TRACE=true cf create-route PROD tcp.ci1-ares-031.swisslab.io --port 31337 | tail -n 31 | jq -r '.resources[].metadata.guid')
# #cf curl /v2/apps/$(cf app minecraft-server-app --guid)
# #cf curl /v2/apps/$(cf app minecraft-server-app --guid) -X PUT -d '{"ports": [8080, 25565]}'
# cf curl /v2/apps/$(cf app minecraft-server-app --guid) -X PUT -d '{"ports": [25565]}'
# #TCP_ROUTE_GUID=$(cf curl /v2/routes | jq -r '.resources[] | select(.entity.port==31337) | .metadata.guid')
# #TCP_ROUTE_GUID=$(cf curl /v2/apps/$(cf app minecraft-server-app --guid)/routes | jq -r '.resources[] | select(.entity.port==31337) | .metadata.guid')
# cf curl /v2/route_mappings -X POST -d "{\"app_guid\": \"$(cf app minecraft-server-app --guid)\", \"route_guid\": \"${TCP_ROUTE_GUID}\", \"app_port\": 25565}"
# #cf curl /v2/apps/$(cf app minecraft-server-app --guid)/route_mappings

# # =============================================================================================
# echo "showing minecraft-server-app routes ..."
# cf routes | grep minecraft-server-app

# # =============================================================================================
# echo "starting minecraft-server-app ..."
# cf start minecraft-server-app
