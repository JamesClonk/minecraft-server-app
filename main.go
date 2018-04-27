package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/JamesClonk/minecraft-server-app/env"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	minio "github.com/minio/minio-go"
)

func main() {
	env, err := cfenv.Current()
	if err != nil {
		log.Fatalf("could not parse VCAP environment: %s\n", err)
	}

	service, err := env.Services.WithName("world-backup")
	if err != nil {
		log.Fatalf("could not get world-backup service from VCAP environment: %s\n", err)
	}

	startServer()
	downloadBackup(service)
}

func startServer() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %s\n", err)
	}
	path := env.MustGet("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s/.java-buildpack/open_jdk_jre/bin", path, pwd))

	cmd := exec.Command("java", "-Xmx1024M", "-Xms1024M", "-jar", "minecraft.jar", "nogui")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		log.Fatalf("starting minecraft server failed: %s\n", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("minecraft server failed: %s\n", err)
	}
}

func downloadBackup(service *cfenv.Service) {
	bucketName := env.MustGet("BUCKET_NAME")
	endpoint, _ := service.CredentialString("accessHost")
	accessKeyID, _ := service.CredentialString("accessKey")
	secretAccessKey, _ := service.CredentialString("sharedSecret")
	useSSL := true

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("could not initialize minio s3 client: %s\n", err)
	}

	if err := minioClient.FGetObject(bucketName, "world-backup.zip", "world-backup.zip", minio.GetObjectOptions{}); err != nil {
		log.Fatalf("could not download world-backup.zip: %s\n", err)
	}
}
