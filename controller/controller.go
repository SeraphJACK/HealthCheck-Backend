package controller

import (
	"errors"
	"flag"
	"fmt"
	"github.com/SeraphJACK/HealthCheck/model"
	"github.com/SeraphJACK/HealthCheck/notify"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var db *gorm.DB

var dbPath = flag.String("db", "data.db", "")
var listen = flag.String("listen", "0.0.0.0:8080", "")

func Start() error {
	var err error
	db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	_ = db.AutoMigrate(&model.Server{})
	_ = db.AutoMigrate(&model.TPS{})

	g := gin.Default()

	registerEndpoints(g)

	go lookAfterServers()

	return http.ListenAndServe(*listen, g)
}

func lookAfterServers() {
	t := time.NewTicker(30 * time.Second)
	for {
		<-t.C
		var servers []model.Server
		db.Where("status = 0").Find(&servers)
		for _, v := range servers {
			var tps model.TPS
			result := db.Order("time desc").Find(&tps, model.TPS{ServerID: v.ID})
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				continue
			}
			if time.Now().Sub(tps.Time) > time.Second*90 {
				v.Status = 2
				db.Save(v)
				notify.Notify(fmt.Sprintf("Server %s lost connection", v.Name), true)
			}
		}
	}
}
