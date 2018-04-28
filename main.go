package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/JamesClonk/minecraft-server-app/env"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/james4k/rcon"
	minio "github.com/minio/minio-go"
)

var (
	s3Client    *minio.Client
	bucketName  = "world-backup"
	backupMutex = &sync.Mutex{}
)

func init() {
	app, err := cfenv.Current()
	if err != nil {
		log.Fatalf("could not parse VCAP environment: %s\n", err)
	}

	service, err := app.Services.WithName("world-backup")
	if err != nil {
		log.Fatalf("could not get world-backup service from VCAP environment: %s\n", err)
	}

	bucketName = env.MustGet("MINECRAFT_BACKUP_BUCKET_NAME")
	endpoint, _ := service.CredentialString("accessHost")
	accessKeyID, _ := service.CredentialString("accessKey")
	secretAccessKey, _ := service.CredentialString("sharedSecret")
	useSSL := true

	s3Client, err = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("could not initialize minio s3 client: %s\n", err)
	}
}

func main() {
	restoreBackup()
	go createBackups()
	startServer()
}

func startServer() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %s\n", err)
	}
	path := env.MustGet("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s/.java-buildpack/open_jdk_jre/bin", path, pwd))

	// update server.properties
	modifyServerProperties("rcon.password", env.MustGet("MINECRAFT_RCON_PASSWORD"))

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

func modifyServerProperties(property, value string) {
	input, err := ioutil.ReadFile("server.properties")
	if err != nil {
		log.Fatalf("could not read server.properties: %s\n", err)
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, property) {
			lines[i] = property + "=" + value
		}
	}

	if err := ioutil.WriteFile("server.properties", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		log.Fatalf("could not write server.properties: %s\n", err)
	}
}

func createBackups() {
	// backup world every 20 minutes, every hour, every weekday, every month, starting immediately after an initial 10min delay
	time.Sleep(10 * time.Minute)
	go backup("world-backup.tar.gz", 20*time.Minute)
	go backup(fmt.Sprintf("world-backup-hour-%s.tar.gz", time.Now().Format("15")), 1*time.Hour)
	go backup(fmt.Sprintf("world-backup-weekday-%s.tar.gz", strings.ToLower(time.Now().Weekday().String())), 24*time.Hour)
	go backup(fmt.Sprintf("world-backup-month-%s.tar.gz", strings.ToLower(time.Now().Month().String())), 31*24*time.Hour)
}

func backup(filename string, interval time.Duration) {
	for {
		backupMutex.Lock()
		fmt.Printf("Starting backup now: [%s] ...\n", filename)
		rconExec("say Starting backup now...")
		rconExec("save-off")
		rconExec("save-all")
		time.Sleep(10 * time.Second)

		os.Remove(filename)
		cmd := exec.Command("tar", "-cvzf", filename, "world/")
		if err := cmd.Run(); err != nil {
			log.Fatalf("could not backup world: %s\n", err)
		}
		rconExec("save-on")

		if _, err := s3Client.FPutObject(bucketName, filename, filename,
			minio.PutObjectOptions{ContentType: "application/gzip"}); err != nil {
			log.Fatalf("could not upload world backup: %s\n", err)
		}

		rconExec("say Backup complete!")
		fmt.Printf("Backup complete! [%s]\n", filename)
		backupMutex.Unlock()

		time.Sleep(interval)
	}
}

func restoreBackup() {
	if err := s3Client.MakeBucket(bucketName, ""); err != nil {
		if exists, err := s3Client.BucketExists(bucketName); err != nil || !exists {
			log.Fatalf("could not create bucket [%s]: %s\n", bucketName, err)
		}
	}

	info, err := s3Client.StatObject(bucketName, "world-backup.tar.gz", minio.StatObjectOptions{})
	if err == nil && info.Size > 0 {
		fmt.Println("Restoring backup...")

		os.Remove("world-backup.tar.gz")
		if err := s3Client.FGetObject(bucketName, "world-backup.tar.gz", "world-backup.tar.gz", minio.GetObjectOptions{}); err != nil {
			log.Fatalf("could not download world backup: %s\n", err)
		}

		if err := os.RemoveAll("world/"); err != nil {
			log.Fatalf("could not delete world folder: %s\n", err)
		}

		//os.Mkdir("world", os.ModePerm)
		cmd := exec.Command("tar", "-xvzf", "world-backup.tar.gz")
		if err := cmd.Run(); err != nil {
			log.Fatalf("could not restore backup world: %s\n", err)
		}
		fmt.Println("Restore complete!")
	}
}

func rconExec(command string) {
	console, err := rcon.Dial("localhost:25575", env.MustGet("MINECRAFT_RCON_PASSWORD"))
	if err != nil {
		log.Fatalf("failed to connect to minecraft rcon server: %s\n", err)
	}
	defer console.Close()

	reqId, err := console.Write(command)
	_, respReqId, err := console.Read()
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Fatalf("failed to read rcon command: %s\n", err)
	}

	if reqId != respReqId {
		fmt.Fprintln(os.Stderr, "response was for another request!")
	}
}
