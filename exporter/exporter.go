package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strconv"
	"upyun-test/httpRequest"
)

const cdnNameSpace = "upYun"

type cdnExporter struct {
	domainList              *[]string
	token                   string
	startTime               string
	endTime                 string
	cdnRequestCount         *prometheus.Desc
	cdnBandWidth            *prometheus.Desc
	cdn4xxErrorRate         *prometheus.Desc
	cdn5xxErrorRate         *prometheus.Desc
	cdnResourceBandWidth    *prometheus.Desc
	cdnResource4xxErrorRate *prometheus.Desc
	cdnResource5xxErrorRate *prometheus.Desc
}

func CdnCloudExporter(domainList *[]string, token string, startTime string, endTime string) *cdnExporter {
	return &cdnExporter{
		domainList: domainList,
		token:      token,
		startTime:  startTime,
		endTime:    endTime,
		cdnRequestCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "request_count"),
			"cdn总请求数",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "bandWidth"),
			"cdn总带宽(Bps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn4xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "4xx_error_rate"),
			"cdn4xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn5xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "5xx_error_rate"),
			"cdn5xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResourceBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_bandWidth"),
			"回源带宽(Bps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource4xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_4xx_error_rate"),
			"cdn回源4xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource5xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_5xx_error_rate"),
			"cdn回源5xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
	}
}
func (e *cdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnRequestCount
	ch <- e.cdnBandWidth
	ch <- e.cdn4xxErrorRate
	ch <- e.cdn5xxErrorRate
	ch <- e.cdnResourceBandWidth
	ch <- e.cdnResource4xxErrorRate
	ch <- e.cdnResource5xxErrorRate
}

func (e *cdnExporter) Collect(ch chan<- prometheus.Metric) {
	for _, domain := range *e.domainList {
		cdnRequestData := httpRequest.DoHttpBandWidthRequest(domain, e.token, e.startTime, e.endTime)
		cdnHealthDegreeData := httpRequest.DoHealthDegreeRequest(e.token, e.startTime, e.endTime).Result
		resourceRequesdata := httpRequest.DoHttpBandWidthResourceRequest(domain, e.token, e.startTime, e.endTime)

		var requestCountTotal float64
		var cdnBandWidthTotal float64
		var http4xxCodeTotal int64
		var http5xxCodeTotal int64
		var ResourceBandWidthCount int
		var ResourceReqsTotal int64
		var ResourceBandWidthTotal float64
		var Resource4xxTotal int64
		var Resource5xxTotal int64
		var Resource4xxErrorAverage float64
		var Resource5xxErrorAverage float64

		for _, point := range cdnRequestData.Data {
			requestCountTotal += point.Reqs
			cdnBandWidthTotal += point.Bandwidth
		}
		for _, v := range resourceRequesdata {
			ResourceBandWidthCount = len(v)
			for _, point := range v {
				Resource4xxTotal += point.Four04
				Resource4xxTotal += point.Four00
				Resource4xxTotal += point.Four03
				Resource4xxTotal += point.Four11
				Resource4xxTotal += point.Four99
				Resource5xxTotal += point.Five03
				Resource5xxTotal += point.Five00
				Resource5xxTotal += point.Five02
				Resource5xxTotal += point.Five04
				ResourceReqsTotal += point.Reqs
				Float64BandWidth, err := strconv.ParseFloat(point.Bandwidth, 64)
				if err != nil {
					log.Fatal(err)
				}
				ResourceBandWidthTotal += Float64BandWidth
			}
		}
		Resource4xxErrorAverage = (float64(Resource4xxTotal) / float64(ResourceReqsTotal)) * 100
		Resource5xxErrorAverage = (float64(Resource5xxTotal) / float64(ResourceReqsTotal)) * 100
		ResourceBandWidthAverage := ResourceBandWidthTotal / float64(ResourceBandWidthCount)

		http4xxCodeTotal = cdnHealthDegreeData.Four99 + cdnHealthDegreeData.Four00 + cdnHealthDegreeData.Four04 + cdnHealthDegreeData.Four03 + cdnHealthDegreeData.Four11
		http5xxCodeTotal = cdnHealthDegreeData.Five00 + cdnHealthDegreeData.Five02 + cdnHealthDegreeData.Five03 + cdnHealthDegreeData.Five04

		http4xxErrorRate := float64(http4xxCodeTotal) / float64(cdnHealthDegreeData.Req)
		http5xxErrorRate := float64(http5xxCodeTotal) / float64(cdnHealthDegreeData.Req)

		cdnBandWidthAverage := cdnBandWidthTotal / float64(len(cdnRequestData.Data))
		requestCountAverage := requestCountTotal / float64(len(cdnRequestData.Data))

		ch <- prometheus.MustNewConstMetric(
			e.cdnRequestCount,
			prometheus.GaugeValue,
			requestCountAverage,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			cdnBandWidthAverage,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdn4xxErrorRate,
			prometheus.GaugeValue,
			http4xxErrorRate,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdn5xxErrorRate,
			prometheus.GaugeValue,
			http5xxErrorRate,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResourceBandWidth,
			prometheus.GaugeValue,
			ResourceBandWidthAverage,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource4xxErrorRate,
			prometheus.GaugeValue,
			Resource4xxErrorAverage,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource5xxErrorRate,
			prometheus.GaugeValue,
			Resource5xxErrorAverage,
			domain,
		)
	}
}

