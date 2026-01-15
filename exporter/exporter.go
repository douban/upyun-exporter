package exporter

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"upyun-exporter/httpRequest"

	"github.com/prometheus/client_golang/prometheus"
)

const cdnNameSpace = "upyun"

func calculateRequestCountPerMin(code float64) float64 {
	return code / 5
}

type CdnExporter struct {
	domainList               *[]string
	token                    string
	rangeTime                int64
	delayTime                int64
	cdnRequestCount          *prometheus.Desc
	cdnResourceRequestCount  *prometheus.Desc
	cdnHitRate               *prometheus.Desc
	cdnFluxHitRate           *prometheus.Desc
	cdnBandWidth             *prometheus.Desc
	cdnResourceBandWidth     *prometheus.Desc
	cdnStatusRate            *prometheus.Desc
	cdnBackSourceStatusRate  *prometheus.Desc
	cdnStatusCount           *prometheus.Desc
	cdnBackSourceStatusCount *prometheus.Desc
	cdnHitTotal              *prometheus.Desc
}

func CdnCloudExporter(domainList *[]string, token string, rangeTime int64, delayTime int64) *CdnExporter {
	return &CdnExporter{
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
		cdnHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "hit_rate"),
			"cdn缓存命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnFluxHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "flux_hit_rate"),
			"cdn缓存字节命中率(%)",
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
			"cdn状态码比例(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
		cdnBackSourceStatusRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_status_rate"),
			"cdn回源状态码比例(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
		cdnStatusCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "status_count"),
			"cdn状态码数量(个)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
		cdnBackSourceStatusCount: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "backsource_status_count"),
			"cdn回源状态码数量(个)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
		cdnHitTotal: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "hit_total"),
			"cdn缓存命中数(次)",
			[]string{
				"instanceId",
			},
			nil,
		),
	}
}

func (e *CdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnRequestCount
	ch <- e.cdnResourceRequestCount
	ch <- e.cdnHitRate
	ch <- e.cdnFluxHitRate
	ch <- e.cdnBandWidth
	ch <- e.cdnResourceBandWidth
	ch <- e.cdnStatusRate
	ch <- e.cdnBackSourceStatusRate
	ch <- e.cdnStatusCount
	ch <- e.cdnBackSourceStatusCount
	ch <- e.cdnHitTotal
}

func (e *CdnExporter) Collect(ch chan<- prometheus.Metric) {
	if len(*e.domainList) == 0 {
		ch <- prometheus.NewInvalidMetric(
			prometheus.NewDesc("upyun_exporter",
				"Error collecting cdn metrics", nil, nil),
			errors.New("empty domain list"))
	}
	var wg sync.WaitGroup
	for _, domain := range *e.domainList {
		domain := domain
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				cdnBandWidthTotal float64
			)
			// 内部有个 判断数据量为 0的逻辑, 所以没法加入 wait group
			// interval - min_five
			cdnRequestData := httpRequest.DoHttpBandWidthRequest(domain, e.token, e.rangeTime, e.delayTime)
			var requestCountTotal float64
			for _, point := range cdnRequestData.Data {
				requestCountTotal += point.Reqs
				cdnBandWidthTotal += point.Bandwidth
			}
			// 去掉数据量为0的数据，得到的结果是NaN
			if requestCountTotal == 0 || cdnBandWidthTotal == 0 {
				return
			}
			requestCountAverage := requestCountTotal / float64(len(cdnRequestData.Data))
			cdnBandWidthAverage := cdnBandWidthTotal / float64(len(cdnRequestData.Data))
			ch <- prometheus.MustNewConstMetric(
				e.cdnRequestCount,
				prometheus.GaugeValue,
				calculateRequestCountPerMin(requestCountAverage),
				domain,
			)
			ch <- prometheus.MustNewConstMetric(
				e.cdnBandWidth,
				prometheus.GaugeValue,
				cdnBandWidthAverage/1000/1000,
				domain,
			)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			cdnFlowDetailData, err := httpRequest.DoHttpFlowDetailRequest(domain, e.token, e.rangeTime, e.delayTime, "cdn")
			if err != nil {
				if err.T == httpRequest.ResponseCodeNot200 {
					ch <- prometheus.NewInvalidMetric(
						prometheus.NewDesc("upyun_exporter",
							"Error collecting cdn flow details", nil, nil),
						err)
					log.Fatalf("error getting metric, return code not 200, error: %s", err)
				} else {
					log.Printf("failed to get cdn flow detail, error: %s", err)
					return
				}
			}

			statusCodes := make(map[string]float64)
			statusCounts := make(map[string]float64)
			var (
				cdnHitRateTotal     float64
				cdnFlowHitRateTotal float64
				codeTotal           int
				code200Total        int
				code206Total        int
				code301Total        int
				code302Total        int
				code304Total        int
				code400Total        int
				code403Total        int
				code404Total        int
				code411Total        int
				code499Total        int
				code500Total        int
				code502Total        int
				code503Total        int
				code504Total        int
				hitTotal            int
			)

			for _, point := range cdnFlowDetailData {
				cdnHitRateTotal = cdnHitRateTotal + (float64(point.Hit))/float64(point.Reqs)
				cdnFlowHitRateTotal = cdnFlowHitRateTotal + (float64(point.HitBytes) / float64(point.Bytes))
				hitTotal += point.Hit
				code200Total += point.Code200
				code206Total += point.Code206
				code301Total += point.Code301
				code302Total += point.Code302
				code304Total += point.Code304
				code400Total += point.Code400
				code403Total += point.Code403
				code404Total += point.Code404
				code411Total += point.Code411
				code499Total += point.Code499
				code500Total += point.Code500
				code502Total += point.Code502
				code503Total += point.Code503
				code504Total += point.Code504
			}
			codeTotal = code200Total + code206Total + code301Total + code302Total + code304Total + code400Total + code403Total +
				code404Total + code411Total + code499Total + code500Total + code502Total + code503Total + code504Total
			statusCodes["200"] = float64(code200Total) / float64(codeTotal)
			statusCodes["206"] = float64(code206Total) / float64(codeTotal)
			statusCodes["2xx"] = float64(code200Total+code206Total) / float64(codeTotal)
			statusCodes["301"] = float64(code301Total) / float64(codeTotal)
			statusCodes["302"] = float64(code302Total) / float64(codeTotal)
			statusCodes["304"] = float64(code304Total) / float64(codeTotal)
			statusCodes["3xx"] = float64(code301Total+code302Total+code304Total) / float64(codeTotal)
			statusCodes["400"] = float64(code400Total) / float64(codeTotal)
			statusCodes["403"] = float64(code403Total) / float64(codeTotal)
			statusCodes["404"] = float64(code404Total) / float64(codeTotal)
			statusCodes["411"] = float64(code411Total) / float64(codeTotal)
			statusCodes["499"] = float64(code499Total) / float64(codeTotal)
			statusCodes["4xx"] = float64(code400Total+code403Total+code404Total+code411Total+code499Total) / float64(codeTotal)
			statusCodes["500"] = float64(code500Total) / float64(codeTotal)
			statusCodes["502"] = float64(code502Total) / float64(codeTotal)
			statusCodes["503"] = float64(code503Total) / float64(codeTotal)
			statusCodes["504"] = float64(code504Total) / float64(codeTotal)
			statusCodes["5xx"] = float64(code500Total+code502Total+code503Total+code504Total) / float64(codeTotal)

			// counts
			statusCounts["200"] = float64(code200Total)
			statusCounts["206"] = float64(code206Total)
			statusCounts["2xx"] = float64(code200Total + code206Total)
			statusCounts["301"] = float64(code301Total)
			statusCounts["302"] = float64(code302Total)
			statusCounts["304"] = float64(code304Total)
			statusCounts["3xx"] = float64(code301Total + code302Total + code304Total)
			statusCounts["400"] = float64(code400Total)
			statusCounts["403"] = float64(code403Total)
			statusCounts["404"] = float64(code404Total)
			statusCounts["411"] = float64(code411Total)
			statusCounts["499"] = float64(code499Total)
			statusCounts["4xx"] = float64(code400Total + code403Total + code404Total + code411Total + code499Total)
			statusCounts["500"] = float64(code500Total)
			statusCounts["502"] = float64(code502Total)
			statusCounts["503"] = float64(code503Total)
			statusCounts["504"] = float64(code504Total)
			statusCounts["5xx"] = float64(code500Total + code502Total + code503Total + code504Total)

			cdnHitRateAverage, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (cdnHitRateTotal/float64(len(cdnFlowDetailData)))*100), 64)
			cdnFlowHitRateAverage, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (cdnFlowHitRateTotal/float64(len(cdnFlowDetailData)))*100), 64)
			ch <- prometheus.MustNewConstMetric(
				e.cdnHitRate,
				prometheus.GaugeValue,
				cdnHitRateAverage,
				domain,
			)
			ch <- prometheus.MustNewConstMetric(
				e.cdnFluxHitRate,
				prometheus.GaugeValue,
				cdnFlowHitRateAverage,
				domain,
			)
			ch <- prometheus.MustNewConstMetric(
				e.cdnHitTotal,
				prometheus.GaugeValue,
				float64(hitTotal),
				domain,
			)
			for status, rate := range statusCodes {
				statusRate, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", rate*100), 64)
				ch <- prometheus.MustNewConstMetric(
					e.cdnStatusRate,
					prometheus.GaugeValue,
					statusRate,
					domain,
					status,
				)
			}

			for status, cnt := range statusCounts {
				ch <- prometheus.MustNewConstMetric(
					e.cdnStatusCount,
					prometheus.GaugeValue,
					cnt,
					domain,
					status,
				)
			}
		}()

		// 回源数据
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				resourceBandwidthTotal float64
				resourceReqsTotal      int
				resourceCodeTotal      int
				resourceCode200Total   int
				resourceCode206Total   int
				resourceCode301Total   int
				resourceCode302Total   int
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

			resourceRequestData, err := httpRequest.DoHttpFlowDetailRequest(domain, e.token, e.rangeTime, e.delayTime, "backsource")
			// response 返回为 {}
			if err != nil {
				return
			}
			for _, point := range resourceRequestData {
				resourceCode200Total += point.Code200
				resourceCode206Total += point.Code206
				resourceCode301Total += point.Code301
				resourceCode302Total += point.Code302
				resourceCode304Total += point.Code304
				resourceCode400Total += point.Code400
				resourceCode403Total += point.Code403
				resourceCode404Total += point.Code404
				resourceCode411Total += point.Code411
				resourceCode499Total += point.Code499
				resourceCode500Total += point.Code500
				resourceCode502Total += point.Code502
				resourceCode503Total += point.Code503
				resourceCode504Total += point.Code504
				resourceBandwidthTotal += point.Bandwidth
				resourceReqsTotal += point.Reqs
			}
			resourceStatusCodes := make(map[string]float64)
			resourceStatusCounts := make(map[string]float64)
			resourceCodeTotal = resourceCode200Total + resourceCode206Total + resourceCode301Total + resourceCode302Total +
				resourceCode304Total + resourceCode400Total + resourceCode403Total + resourceCode404Total + resourceCode411Total +
				resourceCode499Total + resourceCode500Total + resourceCode502Total + resourceCode503Total + resourceCode504Total

			resourceStatusCodes["200"] = float64(resourceCode200Total) / float64(resourceCodeTotal)
			resourceStatusCodes["206"] = float64(resourceCode206Total) / float64(resourceCodeTotal)
			resourceStatusCodes["2xx"] = float64(resourceCode200Total+resourceCode206Total) / float64(resourceCodeTotal)
			resourceStatusCodes["301"] = float64(resourceCode301Total) / float64(resourceCodeTotal)
			resourceStatusCodes["302"] = float64(resourceCode302Total) / float64(resourceCodeTotal)
			resourceStatusCodes["304"] = float64(resourceCode304Total) / float64(resourceCodeTotal)
			resourceStatusCodes["3xx"] = float64(resourceCode301Total+resourceCode302Total+resourceCode304Total) / float64(resourceCodeTotal)
			resourceStatusCodes["400"] = float64(resourceCode400Total) / float64(resourceCodeTotal)
			resourceStatusCodes["403"] = float64(resourceCode403Total) / float64(resourceCodeTotal)
			resourceStatusCodes["404"] = float64(resourceCode404Total) / float64(resourceCodeTotal)
			resourceStatusCodes["411"] = float64(resourceCode411Total) / float64(resourceCodeTotal)
			resourceStatusCodes["499"] = float64(resourceCode499Total) / float64(resourceCodeTotal)
			resourceStatusCodes["4xx"] = float64(resourceCode400Total+resourceCode403Total+resourceCode404Total+resourceCode411Total+resourceCode499Total) / float64(resourceCodeTotal)
			resourceStatusCodes["500"] = float64(resourceCode500Total) / float64(resourceCodeTotal)
			resourceStatusCodes["502"] = float64(resourceCode502Total) / float64(resourceCodeTotal)
			resourceStatusCodes["503"] = float64(resourceCode503Total) / float64(resourceCodeTotal)
			resourceStatusCodes["504"] = float64(resourceCode504Total) / float64(resourceCodeTotal)
			resourceStatusCodes["5xx"] = float64(resourceCode500Total+resourceCode502Total+resourceCode503Total+resourceCode504Total) / float64(resourceCodeTotal)

			// counts
			resourceStatusCounts["200"] = float64(resourceCode200Total)
			resourceStatusCounts["206"] = float64(resourceCode206Total)
			resourceStatusCounts["2xx"] = float64(resourceCode200Total + resourceCode206Total)
			resourceStatusCounts["301"] = float64(resourceCode301Total)
			resourceStatusCounts["302"] = float64(resourceCode302Total)
			resourceStatusCounts["304"] = float64(resourceCode304Total)
			resourceStatusCounts["3xx"] = float64(resourceCode301Total + resourceCode302Total + resourceCode304Total)
			resourceStatusCounts["400"] = float64(resourceCode400Total)
			resourceStatusCounts["403"] = float64(resourceCode403Total)
			resourceStatusCounts["404"] = float64(resourceCode404Total)
			resourceStatusCounts["411"] = float64(resourceCode411Total)
			resourceStatusCounts["499"] = float64(resourceCode499Total)
			resourceStatusCounts["4xx"] = float64(resourceCode400Total + resourceCode403Total + resourceCode404Total + resourceCode411Total + resourceCode499Total)
			resourceStatusCounts["500"] = float64(resourceCode500Total)
			resourceStatusCounts["502"] = float64(resourceCode502Total)
			resourceStatusCounts["503"] = float64(resourceCode503Total)
			resourceStatusCounts["504"] = float64(resourceCode504Total)
			resourceStatusCounts["5xx"] = float64(resourceCode500Total + resourceCode502Total + resourceCode503Total + resourceCode504Total)
			resourceBandwidthAverage := resourceBandwidthTotal / float64(len(resourceRequestData))
			resourceReqsAverage := float64(resourceReqsTotal) / float64(len(resourceRequestData))
			ch <- prometheus.MustNewConstMetric(
				e.cdnResourceBandWidth,
				prometheus.GaugeValue,
				resourceBandwidthAverage/1000/1000,
				domain,
			)

			ch <- prometheus.MustNewConstMetric(
				e.cdnResourceRequestCount,
				prometheus.GaugeValue,
				calculateRequestCountPerMin(resourceReqsAverage),
				domain,
			)
			for status, rate := range resourceStatusCodes {
				statusRate, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", rate*100), 64)
				ch <- prometheus.MustNewConstMetric(
					e.cdnBackSourceStatusRate,
					prometheus.GaugeValue,
					statusRate,
					domain,
					status,
				)
			}

			for status, cnt := range resourceStatusCounts {
				ch <- prometheus.MustNewConstMetric(
					e.cdnBackSourceStatusCount,
					prometheus.GaugeValue,
					cnt,
					domain,
					status,
				)
			}
		}()
	}
	wg.Wait()
}
