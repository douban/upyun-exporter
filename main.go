package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"upyun-test/exporter"
)

var domainList []string

func FetchDomainList() {

	domainList = []string{"img1.doubanio.com"}

}

func main() {
	token := flag.String("token", os.Getenv("UpYun_Token"), "upYun token")
	host := flag.String("host", "0.0.0.0", "服务监听地址")
	port := flag.Int("port", 9300, "服务监听端口")
	delayTime := flag.Int64("delayTime", 60, "时间偏移量, 结束时间=now-delay_seconds")
	rangeTime := flag.Int64("rangeTime", 3000, "选取时间范围, 开始时间=now-range_seconds, 结束时间=now")
	tickerTime := flag.Int("tickerTime", 10,  "刷新域名列表间隔时间")
	flag.Parse()
	FetchDomainList()
	ticker := time.NewTicker(time.Duration(*tickerTime) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				FetchDomainList()
			}
		}
	}()

	cdn := exporter.CdnCloudExporter(&domainList, *token, *rangeTime, *delayTime)
	prometheus.MustRegister(cdn)
	listenAddress := net.JoinHostPort(*host, strconv.Itoa(*port))
	log.Println(listenAddress)
	log.Println("Running on", listenAddress)
	http.Handle("/metrics", promhttp.Handler()) //注册
	log.Fatal(http.ListenAndServe(listenAddress, nil))

}

