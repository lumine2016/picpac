package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type picture struct {
	Id     int64  `db:"id"`
	Sha512 string `db:"sha512"`
	Format string `db:"format"`
}

type Config struct {
	Bind_addr  string `json:"bind_addr"`
	Dsn        string `json:"dsn"`
	Img_dir    string `json:"img_dir"`
	Url_prefix string `json:"url_prefix"`
}

var (
	db     *sqlx.DB
	config Config
)

func initDB() {
	if db == nil {
		dsn := config.Dsn
		newdb, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			log.Panic("connect to database failed\n")
		}
		db = newdb
		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(5)
	}
}

func main() {
	byte, err := os.ReadFile("./config.json")
	if err != nil {
		log.Panic("read config failed\n")
	}
	json.Unmarshal(byte, &config)
	if err != nil {
		log.Panic("read config failed\n")
	}
	initDB()

	r := gin.Default()
	r.Static("/img", config.Img_dir)
	r.GET("/get/:sha512", func(ctx *gin.Context) {
		sha512 := ctx.Param("sha512")
		var pic picture
		err := db.QueryRowx("select * from pictures where upper(sha512) = upper($1)", sha512).StructScan(&pic)
		if err != nil {
			ctx.String(http.StatusAccepted, "not found")
		} else {
			ctx.String(http.StatusAccepted, config.Url_prefix+"/img/"+pic.Sha512+pic.Format)
		}
	})
	r.POST("/upload", func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.Abort()
		}
		format := filepath.Ext(file.Filename)
		if format == "" {
			ctx.Abort()
		}
		multipart_file, err := file.Open()
		if err != nil {
			ctx.Abort()
		}
		h := sha512.New()
		io.Copy(h, multipart_file)
		sha512 := hex.EncodeToString(h.Sum(nil))
		multipart_file.Close()
		var pic picture
		err = db.QueryRowx("select * from pictures where upper(sha512) = upper($1)", sha512).StructScan(&pic)
		if err != nil {
			var last_insert int64
			err := db.QueryRowx("insert into pictures(sha512, format) values($1, $2) returning id", sha512, format).Scan(&last_insert)
			if err != nil {
				ctx.String(http.StatusServiceUnavailable, "")
			}
			err = ctx.SaveUploadedFile(file, path.Join(config.Img_dir, sha512+format))
			if err != nil {
				ctx.String(http.StatusServiceUnavailable, "")
			}
			ctx.String(http.StatusAccepted, sha512)
		} else {
			ctx.String(http.StatusAccepted, sha512)
		}
	})
	r.Run(config.Bind_addr)
}
