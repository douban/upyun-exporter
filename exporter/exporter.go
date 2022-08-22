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
	rangeTime               int64
	delayTime               int64
	cdnRequestCount         *prometheus.Desc
	cdnBandWidth            *prometheus.Desc
	cdn4xxErrorRate         *prometheus.Desc
	cdn5xxErrorRate         *prometheus.Desc
	cdnResourceBandWidth    *prometheus.Desc
	cdnResource4xxErrorRate *prometheus.Desc
	cdnResource5xxErrorRate *prometheus.Desc
	cdnResource4xxCount     *prometheus.Desc
	cdnResource5xxCount     *prometheus.Desc
	cdnResource2xxCount     *prometheus.Desc
	cdnResource3xxCount     *prometheus.Desc
}

func CdnCloudExporter(domainList *[]string, token string, rangeTime int64, delayTime int64) *cdnExporter {
	return &cdnExporter{
		domainList: domainList,
		token:      token,
		rangeTime:  rangeTime,
		delayTime:  delayTime,
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
		cdnResource2xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_2xx_count"),
			"cdn回源2xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource3xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_3xx_count"),
			"cdn回源3xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource4xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_4xx_count"),
			"cdn回源4xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource5xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "resource_5xx_count"),
			"cdn回源5xx请求数(次/分钟)",
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
	ch <- e.cdnResource2xxCount
	ch <- e.cdnResource3xxCount
	ch <- e.cdnResource4xxCount
	ch <- e.cdnResource5xxCount
}

func (e *cdnExporter) Collect(ch chan<- prometheus.Metric) {
	for _, domain := range *e.domainList {
		cdnRequestData := httpRequest.DoHttpBandWidthRequest(domain, e.token, e.rangeTime, e.delayTime)
		cdnHealthDegreeData := httpRequest.DoAccountHealthRequest(e.token, e.rangeTime, e.delayTime).Result
		resourceRequestData := httpRequest.DoHttpBandWidthResourceRequest(domain, e.token, e.rangeTime, e.delayTime)
		var requestCountTotal float64
		var cdnBandWidthTotal float64
		var http4xxCodeTotal int64
		var http5xxCodeTotal int64
		var ResourceBandWidthCount int
		var ResourceReqsTotal int64
		var ResourceBandWidthTotal float64
		var Resource2xxTotal int64
		var Resource3xxTotal int64
		var Resource4xxTotal int64
		var Resource5xxTotal int64
		var Resource2xxTotalAverage float64
		var Resource3xxTotalAverage float64
		var Resource4xxTotalAverage float64
		var Resource5xxTotalAverage float64
		var Resource4xxErrorAverage float64
		var Resource5xxErrorAverage float64

		for _, point := range cdnRequestData.Data {
			requestCountTotal += point.Reqs
			cdnBandWidthTotal += point.Bandwidth
		}
		for _, v := range resourceRequestData {
			ResourceBandWidthCount = len(v)
			for _, point := range v {
				Resource2xxTotal += point.Code200
				Resource2xxTotal += point.Code206
				Resource3xxTotal += point.Code301
				Resource3xxTotal += point.Code303
				Resource3xxTotal += point.Code304
				Resource4xxTotal += point.Code404
				Resource4xxTotal += point.Code400
				Resource4xxTotal += point.Code403
				Resource4xxTotal += point.Code411
				Resource4xxTotal += point.Code499
				Resource5xxTotal += point.Code500
				Resource5xxTotal += point.Code502
				Resource5xxTotal += point.Code503
				Resource5xxTotal += point.Code504
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

		Resource2xxTotalAverage = float64(Resource2xxTotal) / float64(ResourceBandWidthCount)
		Resource3xxTotalAverage = float64(Resource3xxTotal) / float64(ResourceBandWidthCount)
		Resource4xxTotalAverage = float64(Resource4xxTotal) / float64(ResourceBandWidthCount)
		Resource5xxTotalAverage = float64(Resource5xxTotal) / float64(ResourceBandWidthCount)

		http4xxCodeTotal = cdnHealthDegreeData.Code499 + cdnHealthDegreeData.Code400 + cdnHealthDegreeData.Code404 + cdnHealthDegreeData.Code403 + cdnHealthDegreeData.Code411
		http5xxCodeTotal = cdnHealthDegreeData.Code500 + cdnHealthDegreeData.Code502 + cdnHealthDegreeData.Code503 + cdnHealthDegreeData.Code504

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
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource2xxCount,
			prometheus.GaugeValue,
			Resource2xxTotalAverage/5,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource3xxCount,
			prometheus.GaugeValue,
			Resource3xxTotalAverage/5,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource4xxCount,
			prometheus.GaugeValue,
			Resource4xxTotalAverage/5,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource5xxCount,
			prometheus.GaugeValue,
			Resource5xxTotalAverage/5,
			domain,
		)
	}
}

