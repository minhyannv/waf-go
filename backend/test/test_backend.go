package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	// 从环境变量获取端口
	port := 3000
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// 创建HTTP处理器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("收到请求: %s %s", r.Method, r.URL.Path)
		fmt.Fprintf(w, "Hello from backend! Path: %s\n", r.URL.Path)

		// 打印请求头
		fmt.Fprintln(w, "\nRequest Headers:")
		for name, values := range r.Header {
			fmt.Fprintf(w, "%s: %v\n", name, values)
		}
	})

	// 添加健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// 启动HTTP服务器
	log.Printf("后端服务启动在端口 %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
