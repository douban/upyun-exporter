package httpRequest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const httpBandWidthAddress = "https://api.upyun.com/v2/statistics"
const httpAccountHealthAddress = "https://api.upyun.com/flow/health_degree/detail"
const httpBandWidthDetailAddress = "https://api.upyun.com/flow/common_data"

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
type BandWidthDetailList struct {
	Code400   int64  `json:"_400"`
	Code403   int64  `json:"_403"`
	Code404   int64  `json:"_404"`
	Code411   int64  `json:"_411"`
	Code499   int64  `json:"_499"`
	Code500   int64  `json:"_500"`
	Code502   int64  `json:"_502"`
	COde503   int64  `json:"_503"`
	Code504   int64  `json:"_504"`
	Bandwidth string `json:"bandwidth"`
	Reqs      int64  `json:"reqs"`
}
type AccountHealthList struct {
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

func DoAccountHealthRequest(token string, rangeTime int64, delayTime int64) AccountHealthList {
	endTime := time.Now().Add(-time.Minute * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Minute * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpAccountHealthAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	parm := make(url.Values)
	parm.Add("start_time", startTime)
	parm.Add("end_time", endTime)
	req.URL.RawQuery = parm.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	var AccountHealth AccountHealthList
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
	endTime := time.Now().Add(-time.Minute * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Minute * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
	req, err := http.NewRequest("GET", httpBandWidthAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	parm := make(url.Values)
	parm.Add("start_time", startTime)
	parm.Add("end_time", endTime)
	parm.Add("flow_type", "cdn")
	parm.Add("flow_source", "cdn")
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

func DoHttpBandWidthResourceRequest(domain string, token string, rangeTime int64, delayTime int64) map[string][]BandWidthDetailList {
	endTime := time.Now().Add(-time.Minute * time.Duration(delayTime)).Format("2006-01-02:15:04:05")
	startTime := time.Now().Add(-time.Minute * time.Duration(rangeTime)).Format("2006-01-02:15:04:05")
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

	var bandWidthDetail map[string][]BandWidthDetailList
	errJson := json.Unmarshal(body, &bandWidthDetail)
	if errJson != nil {
		log.Fatal(err)
	}
	return bandWidthDetail

}
