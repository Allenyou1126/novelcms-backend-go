package main

import (
	"backend-go/api"
	"backend-go/utils"
	"backend-go/utils/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func Debug() {
	db.EnableDebug()
}

func main() {
	isDebug := func() bool {
		ret := utils.GetConfigWithDefault("DEBUG", "true")
		if ret == "true" || ret == "TRUE" || ret == "True" {
			return true
		}
		return false
	}()
	port, err := strconv.Atoi(utils.GetConfigWithDefault("LISTEN_PORT", "8080"))
	if err != nil || port < 0 || port > 65535 {
		slog.Error("Invalid Environment Variable!")
		os.Exit(1)
	}
	r := gin.New()
	if isDebug {
		gin.SetMode(gin.DebugMode)
		slog.SetDefault(slog.New(utils.NewPrettyHandler(
			os.Stdout,
			utils.PrettyHandlerOptions{
				SlogOpts: slog.HandlerOptions{
					AddSource:   false,
					Level:       slog.LevelDebug,
					ReplaceAttr: nil,
				},
			},
		)))
		corsMiddlewareConfig := cors.DefaultConfig()
		corsMiddlewareConfig.AllowAllOrigins = true
		corsMiddlewareConfig.AddAllowHeaders("Authorization")
		r.Use(func(c *gin.Context) {
			c.Next()
			return
		})
		r.Use(cors.New(corsMiddlewareConfig))
		Debug()
	} else {
		gin.SetMode(gin.ReleaseMode)
		slog.SetDefault(slog.New(utils.NewPrettyHandler(
			os.Stdout,
			utils.PrettyHandlerOptions{
				SlogOpts: slog.HandlerOptions{
					AddSource:   false,
					Level:       slog.LevelInfo,
					ReplaceAttr: nil,
				},
			},
		)))
	}
	r.Use(sloggin.New(slog.Default()))
	api.RegisterRoutes(r)
	err = http.ListenAndServe(":"+strconv.Itoa(port), r)
	if err != nil {
		return
	}
}
