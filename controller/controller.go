package controller

import (
	"errors"
	"flag"
	"github.com/SeraphJACK/HealthCheck/config"
	"github.com/SeraphJACK/HealthCheck/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

var db *gorm.DB

type ServerResponseEntry struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ServerStatus struct {
	LastSeen time.Time `json:"last_seen"`
	Status   uint8     `json:"status"`
	TPS1min  float64   `json:"tps_1_min"`
	TPS5min  float64   `json:"tps_5_min"`
	TPS10min float64   `json:"tps_10_min"`
}

type ServerStatusForm struct {
	Name    string  `json:"name"`
	Last1M  float64 `json:"last1m"`
	Last5M  float64 `json:"last5m"`
	Last10M float64 `json:"last10m"`
}

type LifecycleForm struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func Init() error {
	dbPath := flag.String("db", "data.db", "")
	listen := flag.String("listen", "0.0.0.0:8080", "")
	flag.Parse()

	var err error
	db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	_ = db.AutoMigrate(&model.Server{})
	_ = db.AutoMigrate(&model.TPS{})

	g := gin.Default()

	g.GET("/server", func(ctx *gin.Context) {
		var servers []model.Server
		db.Find(&servers)

		response := make([]ServerResponseEntry, 0)
		for _, v := range servers {
			response = append(response, ServerResponseEntry{ID: v.ID, Name: v.Name})
		}
		ctx.JSON(http.StatusOK, response)
	})

	g.GET("/:server/status", func(ctx *gin.Context) {
		var server model.Server
		s, err := strconv.Atoi(ctx.Param("server"))
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		result := db.First(&server, uint(s))
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		res := ServerStatus{
			Status: server.Status,
		}
		var tps model.TPS
		db.Order("time desc").Where("server_id = ?", server.ID).First(&tps)
		res.LastSeen = tps.Time
		res.TPS1min = tps.Last1m
		res.TPS5min = tps.Last5m
		res.TPS10min = tps.Last10m
		ctx.JSON(http.StatusOK, res)
	})

	g.POST("/status", func(ctx *gin.Context) {
		if ctx.GetHeader("Authorization") != config.Cfg.Token {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		form := ServerStatusForm{}
		if ctx.BindJSON(&form) != nil {
			return
		}
		server := model.Server{}
		db.FirstOrCreate(&server, model.Server{Name: form.Name})
		tps := model.TPS{
			Last1m:  form.Last1M,
			Last5m:  form.Last5M,
			Last10m: form.Last10M,
			Time:    time.Now(),
			Server:  server,
		}
		db.Create(&tps)
		db.Commit()
		ctx.Status(http.StatusNoContent)
	})

	g.POST("/lifecycle", func(ctx *gin.Context) {
		if ctx.GetHeader("Authorization") != config.Cfg.Token {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		form := LifecycleForm{}
		if ctx.BindJSON(&form) != nil {
			return
		}
		server := model.Server{}
		db.FirstOrCreate(&server, model.Server{Name: form.Name})
		if form.Type == "start" {
			server.Status = 0
		} else {
			server.Status = 1
		}
		db.Save(server)
		ctx.Status(http.StatusNoContent)
	})

	return http.ListenAndServe(*listen, g)
}
