package main

import (
	"fmt"
	"log"

	"blog.alphazer01214.top/cmd"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/router"
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

	// 健康检查端点
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	// 设置用户路由
	router.SetupUserRouter(r)
	router.SetupPostRouter(r)

	// 从配置文件读取端口启动服务器
	addr := ":" + global.Config.Server.Port
	fmt.Printf("Starting server on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
