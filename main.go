package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"blog.alphazer01214.top/cmd"
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/keepalive"
	"blog.alphazer01214.top/internal/repository"
	"blog.alphazer01214.top/internal/router"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/internal/utils"
	"blog.alphazer01214.top/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	global.Init()
	if !utils.IsGlobalVarLoaded() {
		panic("not loaded")
	}

	cmd.InitFlag()
	gin.SetMode(global.Config.Server.Mode)

	repos := repository.NewRepositories(global.GetDB())
	svc := service.NewServices(repos)
	apiHandlers := api.NewApis(svc)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	router.SetupUserRouter(r, apiHandlers, svc.UserService)
	router.SetupPostRouter(r, apiHandlers, svc.UserService)
	router.SetupCommentRouter(r, apiHandlers, svc.UserService)
	router.SetupServiceRouter(r, apiHandlers, svc.UserService)
	router.SetupFileRouter(r, apiHandlers, svc.UserService)
	router.SetupTomoriRouter(r, apiHandlers, svc.UserService)
	router.SetupVideoRouter(r, apiHandlers, svc.UserService)

	ka := keepalive.NewManager()
	marketSvc := svc.MarketService
	ka.Register("market_poll", marketSvc.CacheDuration(), marketSvc.PollOnce)
	ka.Start()
	defer ka.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	addr := ":" + global.Config.Server.Port
	fmt.Printf("Starting server on %s\n", addr)

	go func() {
		cert := global.Config.Server.TLSCert
		key := global.Config.Server.TLSKey
		if cert != "" && key != "" {
			fmt.Printf("TLS enabled, cert=%s key=%s\n", cert, key)
			if err := r.RunTLS(addr, cert, key); err != nil {
				log.Fatalf("Failed to start TLS server: %v", err)
			}
		} else {
			if err := r.Run(addr); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	<-quit
	fmt.Println("Shutting down server...")
}
