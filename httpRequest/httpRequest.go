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
const httpAccountHealthAddress = "https://api.upyun.com/flow/health_degree/detail"
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

type BandWidthDetail struct {
	Code200   int  `json:"_200"`
	Code206   int  `json:"_206"`
	Code301   int  `json:"_301"`
	Code303   int  `json:"_303"`
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
	Bandwidth string `json:"bandwidth"`
	Reqs      int  `json:"reqs"`
}

type AccountHealth struct {
	Result struct {
		Code200 int64 `json:"_200"`
		Code400 int64 `json:"_400"`
		Code403 int64 `json:"_403"`
		Code404 int64 `json:"_404"`
		Code411 int64 `json:"_411"`
		Code499 int64 `json:"_499"`
		Code500 int64 `json:"_500"`
		Code502 int64 `json:"_502"`
		Code503 int64 `json:"_503"`
		Code504 int64 `json:"_504"`
		Req     int64 `json:"req"`
	} `json:"result"`
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

func DoAccountHealthRequest(token string, rangeTime int64, delayTime int64) AccountHealth {
	endTime := time.Now().Add(-time.Minute * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Minute * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpAccountHealthAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := make(url.Values)
	params.Add("start_time", startTime)
	params.Add("end_time", endTime)
	params.Add("flow_source", "cdn")
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	var AccountHealth AccountHealth
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	errJson := json.Unmarshal(body, &AccountHealth)
	if errJson != nil {
		log.Fatal(err)
	}
	return AccountHealth
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

func DoHttpBandWidthResourceRequest(domain string, token string, rangeTime int64, delayTime int64) map[string][]BandWidthDetail {
	endTime := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpBandWidthDetailAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	parm := make(url.Values)
	parm.Add("start_time", startTime)
	parm.Add("end_time", endTime)
	//	parm.Add("flow_type", "cdn")
	parm.Add("flow_source", "backsource")
	parm.Add("domain", domain)

	req.URL.RawQuery = parm.Encode()
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
	var bandWidthDetail map[string][]BandWidthDetail
	errJson := json.Unmarshal(body, &bandWidthDetail)
	if errJson != nil {
		log.Fatal(err)
	}
	return bandWidthDetail

}

