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
	rangeTime := flag.Int64("rangeTime", 3000, "upYun rangeTime")
	delayTime := flag.Int64("delayTime", 60, "upYun delayTime")
	tickerTime := flag.Int("tickerTime", 10, "tickerTime")
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
