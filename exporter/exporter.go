package exporter

import (
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
	domainList              *[]string
	token                   string
	rangeTime               int64
	delayTime               int64
	cdnRequestCount         *prometheus.Desc
	cdnBandWidth            *prometheus.Desc
	cdn4xxErrorRate         *prometheus.Desc
	cdn5xxErrorRate         *prometheus.Desc
	cdn2xxCount             *prometheus.Desc
	cdn3xxCount             *prometheus.Desc
	cdn4xxCount             *prometheus.Desc
	cdn5xxCount             *prometheus.Desc
	cdnResourceBandWidth    *prometheus.Desc
	cdnResource4xxErrorRate *prometheus.Desc
	cdnResource5xxErrorRate *prometheus.Desc
	cdnResource4xxCount     *prometheus.Desc
	cdnResource5xxCount     *prometheus.Desc
	cdnResource2xxCount     *prometheus.Desc
	cdnResource3xxCount     *prometheus.Desc
	cdnResourceCodeDetail   *prometheus.GaugeVec
	cdnCodeDetail           *prometheus.GaugeVec
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
			prometheus.BuildFQName(cdnNameSpace, "backsource", "resource_bandWidth"),
			"回源带宽(Bps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource4xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "resource_4xx_error_rate"),
			"cdn回源4xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource5xxErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "resource_5xx_error_rate"),
			"cdn回源5xx错误率",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource2xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "2xx_count"),
			"cdn回源2xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource3xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "3xx_count"),
			"cdn回源3xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource4xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "4xx_count"),
			"cdn回源4xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResource5xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "backsource", "5xx_count"),
			"cdn回源5xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn2xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "2xx_count"),
			"cdn2xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn3xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "3xx_count"),
			"cdn3xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn4xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "4xx_count"),
			"cdn4xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdn5xxCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "5xx_count"),
			"cdn5xx请求数(次/分钟)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnResourceCodeDetail: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cdnNameSpace,
				Subsystem: "backsource",
				Name:      "code_detail",
				Help:      "cdn回源请求数详细分布(次/分钟)",
			},
			[]string{
				"instanceId",
				"code",
			},
		),
		cdnCodeDetail: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: cdnNameSpace,
				Subsystem: "cdn",
				Name:      "code_detail",
				Help:      "cdn请求数详细分布(次/分钟)",
			},
			[]string{
				"instanceId",
				"code",
			},
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
	e.cdnResourceCodeDetail.Describe(ch)
}

func (e *cdnExporter) Collect(ch chan<- prometheus.Metric) {
	for _, domain := range *e.domainList {
		cdnRequestData := httpRequest.DoHttpBandWidthRequest(domain, e.token, e.rangeTime, e.delayTime)
		cdnAccountHealthData := httpRequest.DoAccountHealthRequest(e.token, e.rangeTime, e.delayTime).Result
		resourceRequestData := httpRequest.DoHttpBandWidthResourceRequest(domain, e.token, e.rangeTime, e.delayTime)
		var requestCountTotal float64
		var cdnBandWidthTotal float64
		var http2xxCodeTotal int64
		var http3xxCodeTotal int64
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

		var Code200 int64
		var Code206 int64
		var Code301 int64
		var Code303 int64
		var Code304 int64
		var Code400 int64
		var Code403 int64
		var Code404 int64
		var Code411 int64
		var Code499 int64
		var Code500 int64
		var Code502 int64
		var Code503 int64
		var Code504 int64

		for _, point := range cdnRequestData.Data {
			requestCountTotal += point.Reqs
			cdnBandWidthTotal += point.Bandwidth
		}
		for _, v := range resourceRequestData {
			ResourceBandWidthCount = len(v)
			for _, point := range v {
				Resource2xxTotal += point.Code200
				Code200 = point.Code200
				Resource2xxTotal += point.Code206
				Code206 = point.Code206
				Resource3xxTotal += point.Code301
				Code301 = point.Code301
				Resource3xxTotal += point.Code303
				Code303 = point.Code303
				Resource3xxTotal += point.Code304
				Code304 = point.Code304
				Resource4xxTotal += point.Code404
				Code404 = point.Code404
				Resource4xxTotal += point.Code400
				Code400 = point.Code400
				Resource4xxTotal += point.Code403
				Code403 = point.Code403
				Resource4xxTotal += point.Code411
				Code411 = point.Code411
				Resource4xxTotal += point.Code499
				Code499 = point.Code499
				Resource5xxTotal += point.Code500
				Code500 = point.Code500
				Resource5xxTotal += point.Code502
				Code502 = point.Code502
				Resource5xxTotal += point.Code503
				Code503 = point.Code503
				Resource5xxTotal += point.Code504
				Code504 = point.Code504
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

		http2xxCodeTotal = cdnAccountHealthData.Code200 + cdnAccountHealthData.Code206
		http3xxCodeTotal = cdnAccountHealthData.Code301 + cdnAccountHealthData.Code303 + cdnAccountHealthData.Code304
		http4xxCodeTotal = cdnAccountHealthData.Code499 + cdnAccountHealthData.Code400 + cdnAccountHealthData.Code404 + cdnAccountHealthData.Code403 + cdnAccountHealthData.Code411
		http5xxCodeTotal = cdnAccountHealthData.Code500 + cdnAccountHealthData.Code502 + cdnAccountHealthData.Code503 + cdnAccountHealthData.Code504

		http4xxErrorRate := float64(http4xxCodeTotal) / float64(cdnAccountHealthData.Req)
		http5xxErrorRate := float64(http5xxCodeTotal) / float64(cdnAccountHealthData.Req)

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
			calculateRequestCountPerMin(Resource2xxTotalAverage),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource3xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(Resource3xxTotalAverage),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource4xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(Resource4xxTotalAverage),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnResource5xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(Resource5xxTotalAverage),
			domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn2xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(float64(http2xxCodeTotal)),
			domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn3xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(float64(http3xxCodeTotal)),
			domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn4xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(float64(http4xxCodeTotal)),
			domain,
		)

		ch <- prometheus.MustNewConstMetric(
			e.cdn5xxCount,
			prometheus.GaugeValue,
			calculateRequestCountPerMin(float64(http5xxCodeTotal)),
			domain,
		)

		e.cdnResourceCodeDetail.WithLabelValues(domain, "code200").Set(calculateRequestCountPerMin(float64(Code200)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code206").Set(calculateRequestCountPerMin(float64(Code206)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code301").Set(calculateRequestCountPerMin(float64(Code301)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code303").Set(calculateRequestCountPerMin(float64(Code303)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code304").Set(calculateRequestCountPerMin(float64(Code304)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code400").Set(calculateRequestCountPerMin(float64(Code400)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code403").Set(calculateRequestCountPerMin(float64(Code403)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code404").Set(calculateRequestCountPerMin(float64(Code404)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code411").Set(calculateRequestCountPerMin(float64(Code411)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code499").Set(calculateRequestCountPerMin(float64(Code499)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code500").Set(calculateRequestCountPerMin(float64(Code500)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code502").Set(calculateRequestCountPerMin(float64(Code502)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code503").Set(calculateRequestCountPerMin(float64(Code503)))
		e.cdnResourceCodeDetail.WithLabelValues(domain, "code504").Set(calculateRequestCountPerMin(float64(Code504)))

		e.cdnCodeDetail.WithLabelValues(domain, "code200").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code200)))
		e.cdnCodeDetail.WithLabelValues(domain, "code206").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code206)))
		e.cdnCodeDetail.WithLabelValues(domain, "code301").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code301)))
		e.cdnCodeDetail.WithLabelValues(domain, "code303").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code303)))
		e.cdnCodeDetail.WithLabelValues(domain, "code304").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code304)))
		e.cdnCodeDetail.WithLabelValues(domain, "code400").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code400)))
		e.cdnCodeDetail.WithLabelValues(domain, "code403").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code403)))
		e.cdnCodeDetail.WithLabelValues(domain, "code404").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code404)))
		e.cdnCodeDetail.WithLabelValues(domain, "code411").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code411)))
		e.cdnCodeDetail.WithLabelValues(domain, "code499").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code499)))
		e.cdnCodeDetail.WithLabelValues(domain, "code500").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code500)))
		e.cdnCodeDetail.WithLabelValues(domain, "code502").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code502)))
		e.cdnCodeDetail.WithLabelValues(domain, "code503").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code503)))
		e.cdnCodeDetail.WithLabelValues(domain, "code504").Set(calculateRequestCountPerMin(float64(cdnAccountHealthData.Code504)))

		e.cdnResourceCodeDetail.Collect(ch)
		e.cdnCodeDetail.Collect(ch)

	}
}
