package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"blog.alphazer01214.top/cmd"
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/global"
	internalKafka "blog.alphazer01214.top/internal/kafka"
	"blog.alphazer01214.top/internal/keepalive"
	"blog.alphazer01214.top/internal/router"
	"blog.alphazer01214.top/internal/search"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/internal/utils"
	"blog.alphazer01214.top/pkg/kafka"
	"blog.alphazer01214.top/pkg/middleware"
	ws "blog.alphazer01214.top/pkg/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	global.Init()
	if !utils.IsGlobalVarLoaded() {
		panic("not loaded")
	}

	cmd.InitFlag()
	gin.SetMode(global.Config.Server.Mode)

	// 初始化 Elasticsearch
	searchSvc := initElasticsearch()

	// 初始化 WebSocket Hub
	wsHub := initWebSocket()

	// 初始化 Kafka（依赖 ES 服务和 WebSocket Hub）
	kafkaCleanup := initKafka(searchSvc, wsHub)
	defer kafkaCleanup()

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

	// 设置路由
	router.SetupUserRouter(r)
	router.SetupPostRouter(r)
	router.SetupCommentRouter(r)
	router.SetupServiceRouter(r)
	router.SetupFileRouter(r)
	router.SetupTomoriRouter(r)
	router.SetupVideoRouter(r)
	router.SetupWsRouter(r)
	router.SetupSignalRouter(r)
	router.SetupWatchlistRouter(r)
	router.SetupStockRouter(r)

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
	_ = wsHub // 防止未使用警告
}

// initElasticsearch 初始化 Elasticsearch 搜索服务
func initElasticsearch() *search.SearchService {
	cfg := global.GetConfig()
	if cfg.Elasticsearch == nil || !cfg.Elasticsearch.Enabled || len(cfg.Elasticsearch.Addresses) == 0 {
		fmt.Println("[elasticsearch] disabled or no addresses configured")
		return nil
	}

	searchSvc, err := search.NewSearchService(cfg.Elasticsearch.Addresses, cfg.Elasticsearch.Index)
	if err != nil {
		fmt.Printf("[elasticsearch] init failed: %v\n", err)
		return nil
	}

	// 注入到 service 层
	service.SearchSvc = searchSvc
	fmt.Println("[elasticsearch] initialized")
	return searchSvc
}

// initKafka 初始化 Kafka 生产者和消费者
func initKafka(searchSvc *search.SearchService, wsHub *ws.Hub) func() {
	cfg := global.GetConfig()
	if cfg.Kafka == nil || !cfg.Kafka.Enabled || len(cfg.Kafka.Brokers) == 0 {
		fmt.Println("[kafka] disabled or no brokers configured")
		return func() {}
	}

	brokers := cfg.Kafka.Brokers
	rdb := global.GetRedis()

	// 初始化生产者
	producer := kafka.NewProducer(brokers, kafka.TopicNotification)
	global.KafkaWriter = producer
	fmt.Println("[kafka] producer initialized")

	// 创建消费者（注入 ES 服务和 WebSocket Hub）
	notificationHandler := internalKafka.NewNotificationHandler()
	notificationHandler.SetHub(wsHub) // 注入 WebSocket Hub
	postEventHandler := internalKafka.NewPostEventHandler(searchSvc)
	cacheHandler := internalKafka.NewCacheHandler()

	consumers := []*kafka.Consumer{
		kafka.NewConsumer(brokers, kafka.TopicNotification, "notification-group", notificationHandler.Handle, rdb),
		kafka.NewConsumer(brokers, kafka.TopicPostEvent, "post-event-group", postEventHandler.Handle, rdb),
		kafka.NewConsumer(brokers, kafka.TopicUserEvent, "cache-group", cacheHandler.Handle, rdb),
	}

	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	for _, c := range consumers {
		go c.Consume(ctx)
	}
	fmt.Println("[kafka] consumers started")

	// 返回清理函数
	return func() {
		cancel()
		for _, c := range consumers {
			c.Close()
		}
		producer.Close()
		fmt.Println("[kafka] all stopped")
	}
}

// initWebSocket 初始化 WebSocket Hub
func initWebSocket() *ws.Hub {
	hub := ws.NewHub()
	go hub.Run()

	// 注入到 API 层
	api.WsHub = hub

	// 注入到 Kafka 通知处理器（用于实时推送）
	// 注意：这里需要在 Kafka 消费者启动前完成注入

	fmt.Println("[websocket] hub started")
	return hub
}
