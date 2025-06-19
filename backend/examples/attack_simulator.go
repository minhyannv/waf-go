package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// 攻击类型
var attacks = []struct {
	Name   string
	Path   string
	Params map[string]string
}{
	{
		Name:   "SQL注入",
		Path:   "/?id=1' OR '1'='1",
		Params: map[string]string{},
	},
	{
		Name:   "XSS攻击",
		Path:   "/?q=<script>alert('xss')</script>",
		Params: map[string]string{},
	},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run attack_simulator.go http://目标地址[:端口]")
		fmt.Println("示例: go run attack_simulator.go http://localhost:8080")
		return
	}
	baseURL := os.Args[1]

	for _, atk := range attacks {
		target := baseURL + atk.Path
		fmt.Printf("\n=== 测试: %s ===\n", atk.Name)
		fmt.Printf("请求: %s\n", target)
		resp, err := http.Get(target)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("响应状态: %s\n", resp.Status)
		fmt.Printf("响应内容: %s\n", string(body))
	}

	// 也可以扩展POST等其他攻击
	fmt.Println("\n测试结束。可根据需要自行扩展攻击类型。")
}
