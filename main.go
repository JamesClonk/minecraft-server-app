package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
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
	app         *cfenv.App
)

func init() {
	var err error
	app, err = cfenv.Current()
	if err != nil {
		log.Fatalf("could not parse VCAP environment: %v\n", err)
	}

	service, err := app.Services.WithName("world-backup")
	if err != nil {
		log.Fatalf("could not get world-backup service from VCAP environment: %v\n", err)
	}

	bucketName = env.MustGet("MINECRAFT_BACKUP_BUCKET_NAME")
	endpoint, _ := service.CredentialString("accessHost")
	accessKeyID, _ := service.CredentialString("accessKey")
	secretAccessKey, _ := service.CredentialString("sharedSecret")
	useSSL := true

	s3Client, err = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("could not initialize minio s3 client: %v\n", err)
	}
}

func main() {
	restoreBackup()
	go createBackups()
	go tcpProxy()
	startServer()
}

func tcpProxy() {
	// start listener immediately to pass healthcheck
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Port))
	if err != nil {
		log.Fatalf("could not setup tcp proxy listener: %v\n", err)
	}

	// wait until minecraft server is ready.. 2 minutes should be plenty
	time.Sleep(2 * time.Minute)

	// accept new incoming connections and tcp-forward them to minecraft server
	for {
		//fmt.Println("Awaiting new incoming Minecraft client connection ...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept tcp proxy listener: %v\n", err)
		}
		tcpForward(conn)
	}
}

func tcpForward(conn net.Conn) {
	client, err := net.Dial("tcp", "127.0.0.1:25565")
	if err != nil {
		log.Fatalf("could not connect tcp proxy to minecraft server: %v\n", err)
	}

	//fmt.Printf("TCP proxy connected: %v\n", conn)
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func startServer() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %v\n", err)
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
		log.Fatalf("starting minecraft server failed: %v\n", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("minecraft server failed: %v\n", err)
	}
}

func modifyServerProperties(property, value string) {
	input, err := ioutil.ReadFile("server.properties")
	if err != nil {
		log.Fatalf("could not read server.properties: %v\n", err)
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, property) {
			lines[i] = property + "=" + value
		}
	}

	if err := ioutil.WriteFile("server.properties", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		log.Fatalf("could not write server.properties: %v\n", err)
	}
}

func createBackups() {
	// backup world every 20 minutes, every hour, every weekday, every month, starting immediately after an initial 10min delay
	time.Sleep(10 * time.Minute)

	// every 20 minutes
	go func() {
		for {
			backup("world-backup.tar.gz")
			time.Sleep(20 * time.Minute)
		}
	}()
	// every hour
	go func() {
		for {
			backup(fmt.Sprintf("world-backup-hour-%s.tar.gz", time.Now().Format("15")))
			time.Sleep(1 * time.Hour)
		}
	}()
	// every weekday
	go func() {
		for {
			backup(fmt.Sprintf("world-backup-weekday-%s.tar.gz", strings.ToLower(time.Now().Weekday().String())))
			time.Sleep(24 * time.Hour)
		}
	}()
	// every month
	go func() {
		for {
			backup(fmt.Sprintf("world-backup-month-%s.tar.gz", strings.ToLower(time.Now().Month().String())))
			time.Sleep(30 * 24 * time.Hour)
		}
	}()
}

func backup(filename string) {
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
		log.Fatalf("could not upload world backup: %v\n", err)
	}

	rconExec("say Backup complete!")
	fmt.Printf("Backup complete! [%s]\n", filename)
	backupMutex.Unlock()
}

func restoreBackup() {
	if err := s3Client.MakeBucket(bucketName, ""); err != nil {
		if exists, err := s3Client.BucketExists(bucketName); err != nil || !exists {
			log.Fatalf("could not create bucket [%s]: %v\n", bucketName, err)
		}
	}

	info, err := s3Client.StatObject(bucketName, "world-backup.tar.gz", minio.StatObjectOptions{})
	if err == nil && info.Size > 0 {
		fmt.Println("Restoring backup...")

		os.Remove("world-backup.tar.gz")
		if err := s3Client.FGetObject(bucketName, "world-backup.tar.gz", "world-backup.tar.gz", minio.GetObjectOptions{}); err != nil {
			log.Fatalf("could not download world backup: %v\n", err)
		}

		if err := os.RemoveAll("world/"); err != nil {
			log.Fatalf("could not delete world folder: %v\n", err)
		}

		//os.Mkdir("world", os.ModePerm)
		cmd := exec.Command("tar", "-xvzf", "world-backup.tar.gz")
		if err := cmd.Run(); err != nil {
			log.Fatalf("could not restore backup world: %v\n", err)
		}
		fmt.Println("Restore complete!")
	}
}

func rconExec(command string) {
	console, err := rcon.Dial("localhost:25575", env.MustGet("MINECRAFT_RCON_PASSWORD"))
	if err != nil {
		log.Fatalf("failed to connect to minecraft rcon server: %v\n", err)
	}
	defer console.Close()

	reqId, err := console.Write(command)
	_, respReqId, err := console.Read()
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Fatalf("failed to read rcon command: %v\n", err)
	}

	if reqId != respReqId {
		fmt.Fprintln(os.Stderr, "response was for another request!")
	}
}
