package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/handler"
	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/serviceimpl"
	"github.com/zhangzqs/go-typed-rpc/examples/fullstack/store"
)

func main() {
	// 初始化数据存储
	dataStore := store.GetStore()

	// 创建业务服务实现
	svc := serviceimpl.NewService(dataStore)

	// 设置路由
	r := gin.Default()
	handler.RegisterRouter(r, svc)

	// 启动服务器
	port := "8080"
	baseURL := fmt.Sprintf("http://localhost:%s", port)

	log.Printf("Server starting on %s", baseURL)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
