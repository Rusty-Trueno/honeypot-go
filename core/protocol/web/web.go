package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"honeypot/util/cors"
	"net/http"
	"time"
)

func Start(addr, template, static, url, index string, done chan bool) {
	serverWeb := &http.Server{
		Addr:         addr,
		Handler:      RunWeb(template, index, static, url),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go serverWeb.ListenAndServe()
	go closeServer(serverWeb, done)
}

func RunWeb(template string, index string, static string, url string) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// 引入html资源
	r.LoadHTMLGlob("core/protocol/web/" + template + "/*")

	// 引入静态资源
	r.Static("/static", "core/protocol/web/"+static)

	r.GET(url, func(c *gin.Context) {
		c.HTML(http.StatusOK, index, gin.H{})
	})

	// API 启用状态
	//apiStatus := conf.Get("api", "status")

	// 判断 API 是否启用
	//if apiStatus == "1" {
	// 启动 WEB蜜罐 API
	r.Use(cors.Cors())
	/*webUrl := conf.Get("api", "web_url")
	r.POST(webUrl, api.ReportWeb)*/
	//}

	return r
}

func closeServer(webServer *http.Server, done chan bool) {
	<-done
	fmt.Printf("close server\n")
	if err := webServer.Close(); err != nil {
		fmt.Errorf("http server close failed error is %v\n", err)
	}
}
