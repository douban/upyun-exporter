package httpRequest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const httpBandWidthAddress = "https://api.upyun.com/v2/statistics"
const httpHealthDegreeAddress = "https://api.upyun.com/flow/health_degree/detail"
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
	Four00    int64  `json:"_400"`
	Four03    int64  `json:"_403"`
	Four04    int64  `json:"_404"`
	Four11    int64  `json:"_411"`
	Four99    int64  `json:"_499"`
	Five00    int64  `json:"_500"`
	Five02    int64  `json:"_502"`
	Five03    int64  `json:"_503"`
	Five04    int64  `json:"_504"`
	Bandwidth string `json:"bandwidth"`
	Reqs      int64  `json:"reqs"`
}
type HealthDegreeList struct {
	Result struct {
		Two00  int64 `json:"_200"`
		Four00 int64 `json:"_400"`
		Four03 int64 `json:"_403"`
		Four04 int64 `json:"_404"`
		Four11 int64 `json:"_411"`
		Four99 int64 `json:"_499"`
		Five00 int64 `json:"_500"`
		Five02 int64 `json:"_502"`
		Five03 int64 `json:"_503"`
		Five04 int64 `json:"_504"`
		Req    int64 `json:"req"`
	} `json:"result"`
}

func DoHealthDegreeRequest(token string, startTime string, endTime string) HealthDegreeList {
	req, err := http.NewRequest("GET", httpHealthDegreeAddress, nil)
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
	var healthDegree HealthDegreeList
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	errJson := json.Unmarshal(body, &healthDegree)
	if errJson != nil {
		log.Fatal(err)
	}
	return healthDegree
}
func DoHttpBandWidthRequest(domain string, token string, startTime string, endTime string) BandWidthList {

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

func DoHttpBandWidthResourceRequest(domain string, token string, startTime string, endTime string) map[string][]BandWidthDetailList {
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

