package httpRequest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const domainListAddress = "https://api.upyun.com/buckets"
const httpBandWidthAddress = "https://api.upyun.com/v2/statistics"
const httpBandWidthDetailAddress = "https://api.upyun.com/flow/common_data"

type DomainList struct {
	Domain  string `json:"domain"`
	Status  string `json:"status"`
}

type BucketList struct {
	Buckets []struct {
		BucketId   int64        `json:"bucket_id"`
		BucketName string       `json:"bucket_name"`
		Domains    []DomainList `json:"domains"`
	} `json:"buckets"`
}

type BandWidthList struct {
	Data []struct {
		Bandwidth float64 `json:"bandwidth"`
		Bytes     float64 `json:"bytes"`
		Reqs      float64 `json:"reqs"`
		Rps       float64 `json:"rps"`
		Time      float64 `json:"time"`
	} `json:"data"`
	Interval string `json:"interval"`
}

type FlowDetail struct {
	Code200   int  `json:"_200"`
	Code206   int  `json:"_206"`
	Code301   int  `json:"_301"`
	Code302   int  `json:"_302"`
	Code304   int  `json:"_304"`
	Code400   int  `json:"_400"`
	Code403   int  `json:"_403"`
	Code404   int  `json:"_404"`
	Code411   int  `json:"_411"`
	Code499   int  `json:"_499"`
	Code500   int  `json:"_500"`
	Code502   int  `json:"_502"`
	Code503   int  `json:"_503"`
	Code504   int  `json:"_504"`
	Bandwidth float64 `json:"bandwidth"`
	Reqs      int  `json:"reqs"`
	HitBytes  int  `json:"hit_bytes"`
	Hit       int  `json:"hit"`
	Bytes     int  `json:"bytes"`
}

func DoDomainListRequest(token string) []string {
	req, err := http.NewRequest("GET", domainListAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := make(url.Values)
	params.Add("business_type", "file")
	params.Add("type", "ucdn")
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	var (
		bucketList BucketList
		domainList []string
	)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	errJson := json.Unmarshal(body, &bucketList)
	if errJson != nil {
		log.Fatal(err)
	}

	for _, bucket := range bucketList.Buckets{
		for _, domain := range bucket.Domains{
			if strings.Contains(domain.Domain, "upaiyun") || strings.Contains(domain.Domain, "upcdn") {
				continue
			}
			domainList = append(domainList, domain.Domain)
		}
	}
	return domainList
}

func DoHttpBandWidthRequest(domain string, token string, rangeTime int64, delayTime int64) BandWidthList {
	endTime := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpBandWidthAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	parm := make(url.Values)
	parm.Add("start_time", startTime)
	parm.Add("end_time", endTime)
	parm.Add("flow_type", "cdn")
	parm.Add("flow_source", "backsource")
	parm.Add("domain", domain)
	req.URL.RawQuery = parm.Encode()
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	var BandWidth BandWidthList
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	errJson := json.Unmarshal(body, &BandWidth)
	if errJson != nil {
		log.Fatal(err)
	}
	return BandWidth
}

func DoHttpFlowDetailRequest(domain string, token string, rangeTime int64, delayTime int64, flowSource string) []FlowDetail {
	endTime := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpBandWidthDetailAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := make(url.Values)
	params.Add("start_time", startTime)
	params.Add("end_time", endTime)
	params.Add("query_type", "domain")
	params.Add("query_value", domain)
	params.Add("sum_data", "true")
	// httpcode中不包括200，只有206-504
	if flowSource == "cdn" {
		params.Add("full_region_isp", "true")
		params.Add("fields", "httpcode,hit_bytes,hit,bytes,reqs,_200")
	} else {
		params.Add("flow_source", flowSource)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var detailList []FlowDetail
	errJson := json.Unmarshal(body, &detailList)
	if errJson != nil {
		log.Println(domain, errJson)
		return nil
	}
	return detailList
}

