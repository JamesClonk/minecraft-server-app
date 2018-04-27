FROM openjdk:8-jre

LABEL maintainer "JamesClonk"

# packages
RUN apt-get update && apt-get install -y --no-install-recommends \
		curl \
	&& rm -rf /var/lib/apt/lists/*

# add server
COPY minecraft.jar /server/minecraft.jar
COPY eula.txt /server/eula.txt
COPY server.properties /server/server.properties
COPY server-icon.png /server/server-icon.png
WORKDIR /server

# add supervisor
COPY minecraft-server-app /minecraft-server-app
RUN chmod +x /minecraft-server-app
ENTRYPOINT [ "/minecraft-server-app" ]

# set docker healthcheck
HEALTHCHECK CMD curl -f http://localhost:8080/healthz || exit 1

# ports
EXPOSE 8080 25565 25575

# rcon-cli
ADD https://github.com/itzg/rcon-cli/releases/download/1.3/rcon-cli_linux_amd64 /usr/local/bin/rcon-cli
RUN chmod +x /usr/local/bin/*

# env
ENV JVM_XX_OPTS="-XX:+UseG1GC" MEMORY="1G" \
    LEVEL_TYPE=DEFAULT PVP=true DIFFICULTY=2 GAMEMODE=0 \
    ONLINE_MODE=TRUE CONSOLE=false
