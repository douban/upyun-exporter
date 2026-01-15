package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"upyun-exporter/exporter"
	"upyun-exporter/httpRequest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var domainList []string

func FetchDomainList(token string) {
	domainList = httpRequest.DoDomainListRequest(token)
}

func main() {
	bucketToken := flag.String("bucket_token", os.Getenv("UpYun_Bucket_Token"), "upYun bucket token")
	token := flag.String("token", os.Getenv("UpYun_Token"), "upYun token")
	host := flag.String("host", "0.0.0.0", "服务监听地址")
	port := flag.Int("port", 9300, "服务监听端口")
	delayTime := flag.Int64("delayTime", 300, "时间偏移量, 结束时间=now-delay_seconds")
	rangeTime := flag.Int64("rangeTime", 1800, "选取时间范围, 开始时间=now-range_seconds, 结束时间=now")
	tickerTime := flag.Int("tickerTime", 3600, "刷新域名列表间隔时间")
	metricsPath := flag.String("metricsPath", "/metrics", "默认的metrics路径")
	flag.Parse()
	ticker := time.NewTicker(time.Duration(*tickerTime) * time.Second)
	done := make(chan bool)
	FetchDomainList(*bucketToken)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				FetchDomainList(*bucketToken)
			}
		}
	}()

	cdn := exporter.CdnCloudExporter(&domainList, *token, *rangeTime, *delayTime)
	prometheus.MustRegister(cdn)
	listenAddress := net.JoinHostPort(*host, strconv.Itoa(*port))
	log.Println(listenAddress)
	log.Println("Running on", listenAddress)
	http.Handle(*metricsPath, promhttp.Handler()) //注册

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
           <head><title>UPYUN CDN Exporter</title></head>
           <body>
           <h1>Upyun cdn exporter</h1>
           <p><a href='` + *metricsPath + `'>Metrics</a></p>
           </body>
           </html>`))
	})

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
