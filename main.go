package main

import (
	"flag"
	"fmt"
	"os"

	ldbperm "gitea.com/goftp/leveldb-perm"
	"github.com/lunny/log"
	_ "github.com/shurcooL/vfsgen"
	"github.com/syndtr/goleveldb/leveldb"
	ldbauth "github.com/toaiduongdh/ftpd/modules/ldbauth"
	"github.com/toaiduongdh/ftpd/web"
	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"
	minio_driver "goftp.io/server/v2/driver/minio"
)

var (
	version     = "v0.3.0"
	cfgPath     string
	passivePort string
)

func main() {
	flag.StringVar(&cfgPath, "config", "",
		"config file path, default is config.ini")
	flag.StringVar(&passivePort, "passivePort", "1234-1234",
		"single passive port")
	flag.Parse()

	if err := initConfig(); err != nil {
		fmt.Println(err)
		return
	}

	if err := readConfigFromEnvs(); err != nil {
		fmt.Println(err)
		return
	}

	db, err := leveldb.OpenFile("./authperm.db", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	var auth = &ldbauth.LDBAuth{
		DB: db,
	}
	var perm server.Perm
	if permType == "leveldb" {
		perm = ldbperm.NewLDBPerm(db, "root", "root", os.ModePerm)
	} else {
		perm = server.NewSimplePerm("root", "root")
	}

	var driver server.Driver
	switch driverType {
	case "file":
		_, err = os.Lstat(rootPath)
		if os.IsNotExist(err) {
			os.MkdirAll(rootPath, os.ModePerm)
		} else if err != nil {
			fmt.Println(err)
			return
		}
		driver, err = file.NewDriver(rootPath)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "qiniu":
		/*factory = qiniudriver.NewQiniuDriverFactory(
			qiniu.AccessKey,
			qiniu.SecretKey,
			qiniu.Bucket,
		)*/
	case "minio":
		driver, err = minio_driver.NewDriver(
			minio.Endpoint,
			minio.AccessKey,
			minio.SecretKey,
			"",
			minio.Bucket,
			minio.UseSSL,
		)
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Println("no driver type input")
		return
	}

	// start web manage UI
	if webCfg.Enabled {
		web.DB = auth
		web.Perm = perm
		web.Driver = driver

		go web.Web(webCfg.Listen, "static", "templates", admin, pass,
			webCfg.TLS, webCfg.CertFile, webCfg.KeyFile)
	}

	opt := &server.Options{
		Name:   serv.Name,
		Driver: driver,
		Port:   serv.Port,
		Auth:   auth,
		Perm:   perm,
	}

	opt.TLS = serv.TLS
	opt.KeyFile = serv.KeyFile
	opt.CertFile = serv.CertFile
	opt.ExplicitFTPS = opt.TLS
	opt.PassivePorts = passivePort

	// start ftp server
	ftpServer, err := server.NewServer(opt)
	if err != nil {
		log.Fatal("Error creating server:", err)
	}
	log.Info("FTP Server", version)
	err = ftpServer.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
