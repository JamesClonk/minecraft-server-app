package main

import (
	"log"

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
	downloadBackup(service)
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
