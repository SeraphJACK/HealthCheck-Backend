package controller

import (
	"errors"
	"fmt"
	"github.com/SeraphJACK/HealthCheck/config"
	"github.com/SeraphJACK/HealthCheck/model"
	"github.com/SeraphJACK/HealthCheck/notify"
	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"strconv"
	"time"
)

func NoQuery(ctx *gin.Context) {
	if len(ctx.Request.URL.Query()) != 0 {
		ctx.String(http.StatusForbidden, "query string now allowed")
		ctx.Abort()
	}
}

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
		var tps model.TPS
		db.Order("time desc").Where("server_id = ?", server.ID).First(&tps)
		res := ServerStatus{
			Status:   server.Status,
			LastSeen: tps.Time,
			ServerTPS: ServerTPS{
				TPS1min:     tps.Last1m,
				TPS5min:     tps.Last5m,
				TPS10min:    tps.Last10m,
				PlayerCount: tps.PlayerCount,
			},
		}
		ctx.JSON(http.StatusOK, res)
	})

	g.POST("/status", func(ctx *gin.Context) {
		if ctx.GetHeader("Authorization") != config.Cfg.Token {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		form := ServerStatusForm{}
		if ctx.Bind(&form) != nil {
			return
		}
		server := model.Server{}
		db.FirstOrCreate(&server, model.Server{Name: form.Name})
		if server.Status == 2 {
			notify.Notify(fmt.Sprintf("Server %s back online\n", server.Name), false)
		}
		server.Status = 0
		db.Save(server)
		tps := model.TPS{
			Last1m:      form.Last1M,
			Last5m:      form.Last5M,
			Last10m:     form.Last10M,
			Time:        time.Now(),
			Server:      server,
			ServerID:    server.ID,
			PlayerCount: form.PlayerCount,
		}
		db.Create(&tps)
		if tps.Last1m <= 15 {
			notify.Notify(fmt.Sprintf("Server %s has performance issue, TPS from last 1m, 5m, 10m: %.2f, %.2f, %.2f",
				server.Name, tps.Last1m, tps.Last5m, tps.Last10m), true)
		}
		ctx.Status(http.StatusCreated)
	})

	g.POST("/lifecycle", func(ctx *gin.Context) {
		if ctx.GetHeader("Authorization") != config.Cfg.Token {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		form := LifecycleForm{}
		if ctx.Bind(&form) != nil {
			return
		}
		server := model.Server{}
		db.FirstOrCreate(&server, model.Server{Name: form.Name})
		if form.Type == "start" {
			server.Status = 0
			server.StartTime = time.Now()
			notify.Notify(fmt.Sprintf("Server %s started", form.Name), false)
		} else {
			server.Status = 1
			notify.Notify(fmt.Sprintf("Server %s stopped", form.Name), true)
		}
		db.Save(server)
		ctx.Status(http.StatusNoContent)
	})

	g.GET("/:server/summary", NoQuery, cache.CachePageAtomic(cacheStore, time.Minute, func(ctx *gin.Context) {
		var server model.Server
		s, err := strconv.Atoi(ctx.Param("server"))
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if errors.Is(db.First(&server, uint(s)).Error, gorm.ErrRecordNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		now := time.Now()
		begin := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		res := make([]uint, 24)
		for i := 0; i < 24; i++ {
			var tps model.TPS
			err := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
				Model(&model.TPS{}).
				Where("server_id=?", server.ID).
				Where("time BETWEEN ? AND ?", begin.Add(-time.Duration(i+1)*time.Hour), begin.Add(-time.Duration(i)*time.Hour)).
				Order("last1m").
				First(&tps).Error
			if err != nil {
				// server is down
				res[23-i] = 2
				continue
			}
			if tps.Last1m <= 15.0 {
				// degraded performance
				res[23-i] = 1
			} else {
				// normal
				res[23-i] = 0
			}
		}
		ctx.JSON(http.StatusOK, res)
	}))

	g.GET("/:server/detail", NoQuery, cache.CachePageAtomic(cacheStore, time.Minute, func(ctx *gin.Context) {
		var server model.Server
		s, err := strconv.Atoi(ctx.Param("server"))
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if errors.Is(db.First(&server, uint(s)).Error, gorm.ErrRecordNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		var tps []model.TPS
		db.Where("server_id=?", server.ID).
			Where("time>?", time.Now().Add(-6*time.Hour)).
			Order("time").
			Find(&tps)
		res := make([]ServerTPSDetail, 0, len(tps))
		for _, v := range tps {
			res = append(res, ServerTPSDetail{
				ServerTPS: ServerTPS{
					TPS1min:     v.Last1m,
					TPS5min:     v.Last5m,
					TPS10min:    v.Last10m,
					PlayerCount: v.PlayerCount,
				},
				Time: v.Time.Unix(),
			})
		}
		ctx.JSON(http.StatusOK, res)
	}))
}
