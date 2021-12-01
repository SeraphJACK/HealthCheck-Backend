package model

import (
	"gorm.io/gorm"
	"time"
)

type Server struct {
	gorm.Model
	Name      string `gorm:"unique"`
	StartTime time.Time
	Status    uint8
}

type TPS struct {
	Last1m      float64
	Last5m      float64
	Last10m     float64
	PlayerCount uint
	Time        time.Time
	ServerID    uint
	Server      Server
}
