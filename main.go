package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"blog.alphazer01214.top/cmd"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/keepalive"
	"blog.alphazer01214.top/internal/middleware"
	"blog.alphazer01214.top/internal/router"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	global.Init()
	if !utils.IsGlobalVarLoaded() {
		panic("not loaded")
	}

	cmd.InitFlag()
	gin.SetMode(global.Config.Server.Mode)

	// 创建 Gin 路由实例
	r := gin.Default()

	// 应用 CORS 中间件
	r.Use(middleware.CORSMiddleware())

	// 健康检查端点
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	// 设置用户路由
	router.SetupUserRouter(r)
	router.SetupPostRouter(r)
	router.SetupCommentRouter(r)
	router.SetupServiceRouter(r)
	router.SetupMarketRouter(r)
	router.SetupFileRouter(r)
	router.SetupTomoriRouter(r)

	// keepalive 后台任务
	ka := keepalive.NewManager()

	marketSvc := &service.Service.MarketService
	ka.Register("market_poll", marketSvc.CacheDuration(), marketSvc.PollOnce)
	ka.Start()
	defer ka.Stop()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 从配置文件读取端口启动服务器（支持 HTTP / HTTPS）
	addr := ":" + global.Config.Server.Port
	fmt.Printf("Starting server on %s\n", addr)

	go func() {
		//cert := global.Config.Server.TLSCert
		//key := global.Config.Server.TLSKey
		//if cert != "" && key != "" {
		//	fmt.Printf("TLS enabled, cert=%s key=%s\n", cert, key)
		//	if err := r.RunTLS(addr, cert, key); err != nil {
		//		log.Fatalf("Failed to start TLS server: %v", err)
		//	}
		//} else {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		//}
	}()

	<-quit
	fmt.Println("Shutting down server...")
}
