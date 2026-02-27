package config

import (
	"time"

	"github.com/gin-contrib/cors"
)

func NewCors() cors.Config {
	CorsConf := cors.Config{
		AllowOrigins: []string{
			"https://handworks-cleaning.com",
			"http://localhost:3000",
			// electron renderer process
			"http://localhost:5173", // Electron dev server

		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return CorsConf
}
