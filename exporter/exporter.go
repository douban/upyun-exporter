package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strconv"
	"upyun-test/httpRequest"
)

const cdnNameSpace = "upyun"

func calculateRequestCountPerMin(code float64) float64 {
	return code / 5
}

type cdnExporter struct {
	domainList               *[]string
	token                    string
	rangeTime                int64
	delayTime                int64
	cdnRequestCount          *prometheus.Desc
	cdnResourceRequestCount  *prometheus.Desc
	cdnBandWidth             *prometheus.Desc
	cdnResourceBandWidth     *prometheus.Desc
	cdnStatusRate            *prometheus.Desc
	cdnBackSourceStatusRate  *prometheus.Desc
}

func CdnCloudExporter(domainList *[]string, token string, rangeTime int64, delayTime int64) *cdnExporter {
	return &cdnExporter{
		domainList: domainList,
		token:      token,
		rangeTime:  rangeTime,
		delayTime:  delayTime,

		cdnRequestCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "request_count"),
			"cdn总请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResourceRequestCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_request_count"),
			"cdn回源总请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "bandwidth"),
			"cdn总带宽(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResourceBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "resource_bandwidth"),
			"回源带宽(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnStatusRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "status_rate"),
			"cdn状态码概率(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
		cdnBackSourceStatusRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_status_rate"),
			"cdn回源状态码概率(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
	}
}

func (e *cdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnRequestCount
	ch <- e.cdnResourceRequestCount
	ch <- e.cdnBandWidth
	ch <- e.cdnResourceBandWidth
	ch <- e.cdnStatusRate
	ch <- e.cdnBackSourceStatusRate
}

func (e *cdnExporter) Collect(ch chan<- prometheus.Metric) {
	for _, domain := range *e.domainList {
		var (
			requestCountTotal      float64
			requestCountAverage    float64
			cdnBandWidthTotal      float64
			cdnBandWidthAverage    float64
			resourceBandwidthTotal float64
			resourceBandwidthAverage float64
			resourceReqsTotal      int
			resourceCode200Total   int
			resourceCode206Total   int
			resourceCode301Total   int
			resourceCode303Total   int
			resourceCode304Total   int
			resourceCode400Total   int
			resourceCode403Total   int
			resourceCode404Total   int
			resourceCode411Total   int
			resourceCode499Total   int
			resourceCode500Total   int
			resourceCode502Total   int
			resourceCode503Total   int
			resourceCode504Total   int
		)

		// interval - min_five
		cdnRequestData := httpRequest.DoHttpBandWidthRequest(domain, e.token, e.rangeTime, e.delayTime)
		for _, point := range cdnRequestData.Data {
			requestCountTotal += point.Reqs
			cdnBandWidthTotal += point.Bandwidth
		}
		// 去掉数据量为0的数据，得到的结果是NaN
		if requestCountTotal == 0 || cdnBandWidthTotal == 0 {
			continue
		}
		requestCountAverage = requestCountTotal / float64(len(cdnRequestData.Data))
		cdnBandWidthAverage = cdnBandWidthTotal / float64(len(cdnRequestData.Data))


		resourceRequestData := httpRequest.DoHttpBandWidthResourceRequest(domain, e.token, e.rangeTime, e.delayTime)

		resourceCount := 0
		for _, v := range resourceRequestData {
			for _, point := range v {
				if point.Bandwidth != "" {
					resourceCode200Total += point.Code200
					resourceCode206Total += point.Code206
					resourceCode301Total += point.Code301
					resourceCode303Total += point.Code303
					resourceCode304Total += point.Code304
					resourceCode404Total += point.Code404
					resourceCode400Total += point.Code400
					resourceCode403Total += point.Code403
					resourceCode411Total += point.Code411
					resourceCode499Total += point.Code499
					resourceCode500Total += point.Code500
					resourceCode502Total += point.Code502
					resourceCode503Total += point.Code503
					resourceCode504Total += point.Code504
					resourceReqsTotal += point.Reqs
					Float64BandWidth, err := strconv.ParseFloat(point.Bandwidth, 64)
					if err != nil {
						log.Fatal(err)
					}
					resourceBandwidthTotal += Float64BandWidth
					resourceCount += 1
				}
			}
		}

		if resourceBandwidthTotal == 0 {
			continue
		}
		resourceStatusCodes := make(map[string]float64)
		resourceStatusCodes["200"] = float64(resourceCode200Total) / float64(resourceReqsTotal)
		resourceStatusCodes["206"] = float64(resourceCode206Total) / float64(resourceReqsTotal)
		resourceStatusCodes["2xx"] = float64(resourceCode200Total + resourceCode206Total) / float64(resourceReqsTotal)
		resourceStatusCodes["301"] = float64(resourceCode301Total) / float64(resourceReqsTotal)
		resourceStatusCodes["303"] = float64(resourceCode303Total) / float64(resourceReqsTotal)
		resourceStatusCodes["304"] = float64(resourceCode304Total) / float64(resourceReqsTotal)
		resourceStatusCodes["3xx"] = float64(resourceCode301Total + resourceCode303Total + resourceCode304Total) / float64(resourceReqsTotal)
		resourceStatusCodes["400"] = float64(resourceCode400Total) / float64(resourceReqsTotal)
		resourceStatusCodes["403"] = float64(resourceCode403Total) / float64(resourceReqsTotal)
		resourceStatusCodes["404"] = float64(resourceCode404Total) / float64(resourceReqsTotal)
		resourceStatusCodes["411"] = float64(resourceCode411Total) / float64(resourceReqsTotal)
		resourceStatusCodes["499"] = float64(resourceCode499Total) / float64(resourceReqsTotal)
		resourceStatusCodes["4xx"] = float64(resourceCode400Total + resourceCode403Total + resourceCode404Total + resourceCode411Total + resourceCode499Total) / float64(resourceReqsTotal)
		resourceStatusCodes["500"] = float64(resourceCode500Total) / float64(resourceReqsTotal)
		resourceStatusCodes["502"] = float64(resourceCode502Total) / float64(resourceReqsTotal)
		resourceStatusCodes["503"] = float64(resourceCode503Total) / float64(resourceReqsTotal)
		resourceStatusCodes["504"] = float64(resourceCode504Total) / float64(resourceReqsTotal)
		resourceStatusCodes["5xx"] = float64(resourceCode500Total + resourceCode502Total + resourceCode503Total + resourceCode504Total) / float64(resourceReqsTotal)
		resourceBandwidthAverage = resourceBandwidthTotal / float64(resourceCount)
		resourceReqsAverage := float64(resourceReqsTotal) / float64(resourceCount)

		cdnAccountHealthData := httpRequest.DoAccountHealthRequest(e.token, e.rangeTime, e.delayTime).Result
		statusCodes := make(map[string]float64)
		statusCodes["200"] = float64(cdnAccountHealthData.Code200) / float64(cdnAccountHealthData.Req)
		statusCodes["400"] = float64(cdnAccountHealthData.Code400) / float64(cdnAccountHealthData.Req)
		statusCodes["403"] = float64(cdnAccountHealthData.Code403) / float64(cdnAccountHealthData.Req)
		statusCodes["404"] = float64(cdnAccountHealthData.Code404) / float64(cdnAccountHealthData.Req)
		statusCodes["411"] = float64(cdnAccountHealthData.Code411) / float64(cdnAccountHealthData.Req)
		statusCodes["499"] = float64(cdnAccountHealthData.Code499) / float64(cdnAccountHealthData.Req)
		statusCodes["4xx"] = float64(cdnAccountHealthData.Code400 + cdnAccountHealthData.Code403 + cdnAccountHealthData.Code404 + cdnAccountHealthData.Code411 + cdnAccountHealthData.Code499) / float64(cdnAccountHealthData.Req)
		statusCodes["500"] = float64(cdnAccountHealthData.Code500) / float64(cdnAccountHealthData.Req)
		statusCodes["502"] = float64(cdnAccountHealthData.Code502) / float64(cdnAccountHealthData.Req)
		statusCodes["503"] = float64(cdnAccountHealthData.Code503) / float64(cdnAccountHealthData.Req)
		statusCodes["504"] = float64(cdnAccountHealthData.Code504) / float64(cdnAccountHealthData.Req)
		statusCodes["5xx"] = float64(cdnAccountHealthData.Code500 + cdnAccountHealthData.Code502 + cdnAccountHealthData.Code503 + cdnAccountHealthData.Code504) / float64(cdnAccountHealthData.Req)

		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(requestCountAverage),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResourceRequestCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(resourceReqsAverage),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			cdnBandWidthAverage / 1024 / 1024,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResourceBandWidth,
			prometheus.GaugeValue,
			resourceBandwidthAverage / 1024 / 1024,
			domain,
		)
		for status, rate := range resourceStatusCodes {
			statusRate, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", rate * 100), 64)
			ch <- prometheus.MustNewConstMetric(
				e.cdnBackSourceStatusRate,
				prometheus.GaugeValue,
				statusRate,
				domain,
				status,
			)
		}
		for status, rate := range statusCodes {
			statusRate, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", rate * 100), 64)
			ch <- prometheus.MustNewConstMetric(
				e.cdnStatusRate,
				prometheus.GaugeValue,
				statusRate,
				domain,
				status,
			)
		}
	}
}
