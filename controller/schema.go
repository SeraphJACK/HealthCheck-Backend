package controller

import "time"

type ServerResponseEntry struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ServerTPS struct {
	TPS1min     float64 `json:"tps_1_min"`
	TPS5min     float64 `json:"tps_5_min"`
	TPS10min    float64 `json:"tps_10_min"`
	PlayerCount uint    `json:"player_count"`
}

type ServerTPSDetail struct {
	ServerTPS
	Time int64 `json:"time"`
}

type ServerStatus struct {
	ServerTPS
	LastSeen time.Time `json:"last_seen"`
	Status   uint8     `json:"status"`
}

type ServerStatusForm struct {
	Name        string  `json:"name"`
	Last1M      float64 `json:"last1m"`
	Last5M      float64 `json:"last5m"`
	Last10M     float64 `json:"last10m"`
	PlayerCount uint    `json:"player_count"`
}

type LifecycleForm struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
