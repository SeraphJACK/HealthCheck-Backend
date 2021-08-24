package model

import (
	"gorm.io/gorm"
	"time"
)

type Server struct {
	gorm.Model
	Name   string `gorm:"unique"`
	Status uint8
}

type TPS struct {
	Last1m   float64
	Last5m   float64
	Last10m  float64
	Time     time.Time
	ServerID int
	Server   Server
}
