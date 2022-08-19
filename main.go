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
	startTime := flag.String("startTime", os.Getenv("UpYun_startTime"), "upYun startTime")
	endTime := flag.String("endTime", os.Getenv("UpYun_endTime"), "upYun endTime")
	host := flag.String("host", "0.0.0.0", "服务监听地址")
	port := flag.Int("port", 9200, "服务监听端口")
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
	cdn := exporter.CdnCloudExporter(&domainList, *token, *startTime, *endTime)
	prometheus.MustRegister(cdn)
	listenAddress := net.JoinHostPort(*host, strconv.Itoa(*port))
	log.Println(listenAddress)
	log.Println("Running on", listenAddress)
	http.Handle("/metrics", promhttp.Handler()) //注册
	log.Fatal(http.ListenAndServe(listenAddress, nil))

}

