package controller

import (
	"errors"
	"fmt"
	"github.com/SeraphJACK/HealthCheck/config"
	"github.com/SeraphJACK/HealthCheck/model"
	"github.com/SeraphJACK/HealthCheck/notify"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func registerEndpoints(g *gin.Engine) {
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
		server.Status = 0
		db.Save(server)
		tps := model.TPS{
			Last1m:   form.Last1M,
			Last5m:   form.Last5M,
			Last10m:  form.Last10M,
			Time:     time.Now(),
			Server:   server,
			ServerID: server.ID,
		}
		db.Create(&tps)
		if tps.Last1m <= 15 {
			notify.Notify(fmt.Sprintf("Server %s has performance issue, TPS from last 1m, 5m, 10m: %.2f, %.2f, %.2f",
				server.Name, tps.Last1m, tps.Last5m, tps.Last10m), true)
		}
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
			notify.Notify(fmt.Sprintf("Server %s started", form.Name), false)
		} else {
			server.Status = 1
		}
		db.Save(server)
		ctx.Status(http.StatusNoContent)
	})
}
