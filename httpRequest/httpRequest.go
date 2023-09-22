package httpRequest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	domainListAddress                       = "https://api.upyun.com/buckets"
	httpBandWidthAddress                    = "https://api.upyun.com/v2/statistics"
	httpBandWidthDetailAddress              = "https://api.upyun.com/flow/common_data"
	ParseError                 ApiErrorType = iota
	ResponseCodeNot200
)

type DomainList struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
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

type ApiErrorType uint8

type ApiError struct {
	Message string
	T       ApiErrorType
}

func (r *ApiError) Error() string {
	return fmt.Sprintf("code: %v, message: %s", r.T, r.Message)
}

func NewRequestError(message string, t ApiErrorType) *ApiError {
	return &ApiError{
		Message: message,
		T:       t,
	}
}

type FlowDetail struct {
	Code200   int     `json:"_200"`
	Code206   int     `json:"_206"`
	Code301   int     `json:"_301"`
	Code302   int     `json:"_302"`
	Code304   int     `json:"_304"`
	Code400   int     `json:"_400"`
	Code403   int     `json:"_403"`
	Code404   int     `json:"_404"`
	Code411   int     `json:"_411"`
	Code499   int     `json:"_499"`
	Code500   int     `json:"_500"`
	Code502   int     `json:"_502"`
	Code503   int     `json:"_503"`
	Code504   int     `json:"_504"`
	Bandwidth float64 `json:"bandwidth"`
	Reqs      int     `json:"reqs"`
	HitBytes  int     `json:"hit_bytes"`
	Hit       int     `json:"hit"`
	Bytes     int     `json:"bytes"`
}

type BucketInfo struct {
	BucketName         string        `json:"bucket_name,omitempty"`
	Type               string        `json:"type,omitempty"`
	BusinessType       string        `json:"business_type,omitempty"`
	Status             string        `json:"status,omitempty"`
	Separator          string        `json:"separator,omitempty"`
	Visible            bool          `json:"visible,omitempty"`
	FormApiSecret      string        `json:"form_api_secret,omitempty"`
	CreatedAt          string        `json:"created_at,omitempty"`
	BucketRemark       string        `json:"bucket_remark,omitempty"`
	ApprovalDomains    []string      `json:"approval_domains,omitempty"`
	DisapprovalDomains []interface{} `json:"disapproval_domains,omitempty"`
	ApprovalingDomains []interface{} `json:"approvaling_domains,omitempty"`
	PurgeControl       int           `json:"purge_control,omitempty"`
	DefaultDomain      struct {
		Https      bool   `json:"https,omitempty"`
		ForceHttps bool   `json:"force_https,omitempty"`
		Domain     string `json:"domain,omitempty"`
	} `json:"default_domain,omitempty"`
	Operators        []string `json:"operators,omitempty"`
	FusionCdn        bool     `json:"fusion_cdn,omitempty"`
	SecurityCdn      bool     `json:"security_cdn,omitempty"`
	Websocket        bool     `json:"websocket,omitempty"`
	InfrequentAccess bool     `json:"infrequent_access,omitempty"`
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}
	if response.StatusCode != 200 {
		log.Printf("request: %v, response: %v", req, response)
		log.Fatalf("get domain list failed, return code not 200, response code: %v, response body: %s", response.StatusCode, string(body))
	}

	var (
		bucketList BucketList
		domainList []string
	)

	err = json.Unmarshal(body, &bucketList)
	if err != nil {
		log.Fatal(err)
	}

	for _, bucket := range bucketList.Buckets {
		bucketInfo := GetBucketInfo(bucket.BucketName, token)
		if !bucketInfo.Visible {
			continue
		}
		for _, domain := range bucket.Domains {
			if strings.Contains(domain.Domain, "upaiyun") || strings.Contains(domain.Domain, "upcdn") {
				continue
			}
			domainList = append(domainList, domain.Domain)
		}
	}
	return domainList
}

func GetBucketInfo(domain string, token string) BucketInfo {
	req, err := http.NewRequest("GET", "https://api.upyun.com/buckets/info", nil)
	if err != nil {
		log.Fatal(err)
	}
	params := make(url.Values)
	params.Add("bucket_name", domain)
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	response, err := client.Do(req)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("请求失败", err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}
	if response.StatusCode != 200 {
		log.Printf("request: %v, response: %v", req, response)
		log.Fatalf("get domain info failed, return code not 200, response code: %v, response body: %s", response.StatusCode, string(body))
	}

	var bucketInfo BucketInfo

	err = json.Unmarshal(body, &bucketInfo)
	if err != nil {
		log.Fatal(err)
	}
	return bucketInfo
}

func DoHttpBandWidthRequest(domain string, token string, rangeTime int64, delayTime int64) BandWidthList {
	timeZone, _ := time.LoadLocation("Asia/Shanghai")
	timeNow := time.Now().In(timeZone)
	endTime := timeNow.Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02 15:04:05")
	startTime := timeNow.Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02 15:04:05")
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Fatalf("Failed to get bandwidth data, response: %s", string(body))
	}
	err = json.Unmarshal(body, &BandWidth)
	if err != nil {
		log.Fatal(err)
	}
	return BandWidth
}

func DoHttpFlowDetailRequest(domain string, token string, rangeTime int64, delayTime int64, flowSource string) ([]FlowDetail, *ApiError) {
	timeZone, _ := time.LoadLocation("Asia/Shanghai")
	timeNow := time.Now().In(timeZone)
	endTime := timeNow.Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02 15:04:05")
	startTime := timeNow.Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02 15:04:05")
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}
	if response.StatusCode != 200 {
		log.Printf("request: %v", req)
		return nil, NewRequestError(fmt.Sprintf("get flow data failed, return code not 200, response body: %s",
			string(body)), ResponseCodeNot200)
	}

	var detailList []FlowDetail
	err = json.Unmarshal(body, &detailList)
	if err != nil {
		log.Printf("failed to collect domain flow detail, domain: %s, response: %s", domain, string(body))
		return nil, NewRequestError(fmt.Sprintf("Failed to decode body to flow detail, domain: %s, response: %s, error: %v",
			domain, string(body), err), ParseError)

	}
	return detailList, nil
}
