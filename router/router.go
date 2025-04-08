package router

import (
	"log/slog"

	"github.com/csBenClarkson/url-shortener/store"
	"github.com/gin-gonic/gin"
	"github.com/samber/slog-gin"
)

type storageModel struct {
	storage *store.Storage
}

func SetupRouter(logger *slog.Logger, storage *store.Storage) *gin.Engine {
	engine := gin.New()
	engine.Use(sloggin.New(logger))
	model := storageModel{storage}
	engine.LoadHTMLGlob("router/templates/*")
	registerRoutes(engine, model)
	return engine
}

func registerRoutes(engine *gin.Engine, model storageModel) {
	engine.GET("/", getIndex)
	engine.GET("/:digest", model.getURLEndpoint)
	engine.GET("/admin/login", getLogin)
	engine.POST("/admin/login", postLogin)
	engine.GET("/admin/register", getRegisterURL)
	engine.POST("/admin/register", model.postRegisterURL)
}
