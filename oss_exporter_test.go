package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/jarcoal/httpmock"
)

var (
	testCases = ossExporterTestCases{
		// Test one object in a bucket
		ossExporterTestCase{
			Name:   "one object",
			Bucket: "mock",
			Prefix: "one",
			ExpectedOutputLines: []string{
				"oss_list_success{bucket=\"mock\",prefix=\"one\"} 1",
				"oss_last_modified_object_date{bucket=\"mock\",prefix=\"one\"} 1.5604596e+09",
				"oss_last_modified_object_size_bytes{bucket=\"mock\",prefix=\"one\"} 1234",
				"oss_biggest_object_size_bytes{bucket=\"mock\",prefix=\"one\"} 1234",
				"oss_objects_size_sum_bytes{bucket=\"mock\",prefix=\"one\"} 1234",
				"oss_objects_total{bucket=\"mock\",prefix=\"one\"} 1",
			},
			ListObjectsResult: &oss.ListObjectsResult{
				Objects: []oss.ObjectProperties{
					oss.ObjectProperties{
						Key:          "one",
						LastModified: time.Date(2019, time.June, 13, 21, 0, 0, 0, time.UTC),
						Size:         1234,
					},
				},
				IsTruncated: false,
				// KeyCount:    Int64(1),
				MaxKeys: 1000,
				// XMLName:     "mock",
				Prefix: "one",
			},
		},
		// Test no matching objects in the bucket
		ossExporterTestCase{
			Name:   "no objects",
			Bucket: "mock",
			Prefix: "none",
			ExpectedOutputLines: []string{
				"oss_biggest_object_size_bytes{bucket=\"mock\",prefix=\"none\"} 0",
				"oss_last_modified_object_date{bucket=\"mock\",prefix=\"none\"} -6.795364578e+09",
				"oss_last_modified_object_size_bytes{bucket=\"mock\",prefix=\"none\"} 0",
				"oss_list_success{bucket=\"mock\",prefix=\"none\"} 1",
				"oss_objects_size_sum_bytes{bucket=\"mock\",prefix=\"none\"} 0",
				"oss_objects_total{bucket=\"mock\",prefix=\"none\"} 0",
			},
			ListObjectsResult: &oss.ListObjectsResult{
				Objects:     []oss.ObjectProperties{},
				IsTruncated: false,
				// KeyCount:    Int64(0),
				MaxKeys: 1000,
				// XMLName:     "mock",
				Prefix: "none",
			},
		},
		// Test multiple objects
		ossExporterTestCase{
			Name:   "multiple objects",
			Bucket: "mock",
			Prefix: "multiple",
			ExpectedOutputLines: []string{
				"oss_biggest_object_size_bytes{bucket=\"mock\",prefix=\"multiple\"} 4567",
				"oss_last_modified_object_date{bucket=\"mock\",prefix=\"multiple\"} 1.568592e+09",
				"oss_last_modified_object_size_bytes{bucket=\"mock\",prefix=\"multiple\"} 4567",
				"oss_list_success{bucket=\"mock\",prefix=\"multiple\"} 1",
				"oss_objects_size_sum_bytes{bucket=\"mock\",prefix=\"multiple\"} 11602",
				"oss_objects_total{bucket=\"mock\",prefix=\"multiple\"} 4",
			},
			ListObjectsResult: &oss.ListObjectsResult{
				Objects: []oss.ObjectProperties{
					oss.ObjectProperties{
						Key:          "multiple0",
						LastModified: time.Date(2019, time.June, 13, 21, 0, 0, 0, time.UTC),
						Size:         1234,
					},
					oss.ObjectProperties{
						Key:          "multiple1",
						LastModified: time.Date(2019, time.July, 14, 22, 0, 0, 0, time.UTC),
						Size:         2345,
					},
					oss.ObjectProperties{
						Key:          "multiple2",
						LastModified: time.Date(2019, time.August, 15, 23, 0, 0, 0, time.UTC),
						Size:         3456,
					},
					oss.ObjectProperties{
						Key:          "multiple/0",
						LastModified: time.Date(2019, time.September, 16, 00, 0, 0, 0, time.UTC),
						Size:         4567,
					},
				},
				IsTruncated: false,
				MaxKeys:     1000,
				// XMLName:     String("mock"),
				Prefix: "multiple",
			},
		},
	}
)

type ossExporterTestCase struct {
	Name                string
	Bucket              string
	Prefix              string
	ExpectedOutputLines []string
	ListObjectsResult   *oss.ListObjectsResult
}

// testBody tests the body returned by the exporter against the expected output
func (tc ossExporterTestCase) testBody(body string, t *testing.T) {
	for _, l := range tc.ExpectedOutputLines {
		ok := strings.Contains(body, l)
		if !ok {
			t.Errorf("expected " + l)
		}
	}
}

type ossExporterTestCases []ossExporterTestCase

// TestProbeHandler iterates over a list of test cases
func TestProbeHandler(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	var uri string

	for _, c := range testCases {
		prefix := c.Prefix
		bucket := c.Bucket

		if len(prefix) > 0 {
			uri = "/probe?bucket=" + bucket + "&prefix=" + prefix
		} else {
			uri = "/probe?bucket=" + bucket
		}
		httpmock.RegisterResponder("GET", uri,
			func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, strings.Join(c.ExpectedOutputLines, ","))
				return resp, nil
			},
		)

		resp, err := http.Get(uri)
		if err != nil {
			t.Errorf(err.Error())
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf(err.Error())
		}
		bodyString := string(bodyBytes)
		c.testBody(bodyString, t)
	}
}
