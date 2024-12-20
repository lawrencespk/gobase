package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PrometheusMetric 表示从Prometheus查询到的指标
type PrometheusMetric struct {
	Labels map[string]string // 指标标签
	Value  float64           // 指标值
}

// PrometheusResponse 表示Prometheus API的响应结构
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // 使用 interface{} 来处理混合类型
		} `json:"result"`
	} `json:"data"`
}

// QueryPrometheusMetrics 查询Prometheus指标，带重试机制
func QueryPrometheusMetrics(uri string, query string) ([]PrometheusMetric, error) {
	// 最多重试3次，每次等待1秒
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		metrics, err := queryPrometheusOnce(uri, query)
		if err != nil {
			return nil, err
		}

		// 如果找到指标，直接返回
		if len(metrics) > 0 {
			return metrics, nil
		}

		// 如果没有找到指标且还有重试次数，等待1秒后重试
		if i < maxRetries-1 {
			time.Sleep(time.Second)
		}
	}

	// 所有重试都失败后，返回空结果
	return nil, nil
}

// queryPrometheusOnce 执行单次Prometheus查询
func queryPrometheusOnce(uri string, query string) ([]PrometheusMetric, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query", uri)

	// 先尝试获取所有可用的指标
	metricsURL := fmt.Sprintf("%s/metrics", uri)
	metricsResp, err := http.Get(metricsURL)
	if err == nil {
		body, _ := io.ReadAll(metricsResp.Body)
		metricsResp.Body.Close()
		fmt.Printf("Available metrics:\n%s\n", string(body))
	}

	params := url.Values{}
	params.Add("query", query)

	fullURL := queryURL + "?" + params.Encode()
	fmt.Printf("Querying Prometheus: %s\n", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("查询Prometheus失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体并打印
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Printf("Prometheus响应: %s\n", string(body))

	// 尝试直接解析为map来查看结构
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		fmt.Printf("Raw response parse error: %v\n", err)
	} else {
		fmt.Printf("Raw response structure: %+v\n", rawResp)
	}

	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	var metrics []PrometheusMetric
	for _, result := range promResp.Data.Result {
		if len(result.Value) != 2 {
			continue
		}

		// 第二个元素是值，可能是字符串或数字
		var value float64
		switch v := result.Value[1].(type) {
		case string:
			val, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("解析字符串值失败: %v", err)
			}
			value = val
		case float64:
			value = v
		case json.Number:
			val, err := v.Float64()
			if err != nil {
				return nil, fmt.Errorf("解析数字值失败: %v", err)
			}
			value = val
		default:
			return nil, fmt.Errorf("未知的值类型: %T, 值: %v", v, v)
		}

		metrics = append(metrics, PrometheusMetric{
			Labels: result.Metric,
			Value:  value,
		})
	}

	return metrics, nil
}
