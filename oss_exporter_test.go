package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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

// Returns the mocked response for a bucket+prefix combination
//func (tcs *ossExporterTestCases) response(bucket, prefix string) (*oss.ListObjectsResult, error) {
//	for _, c := range *tcs {
//		if c.Bucket == bucket && c.Prefix == prefix {
//			return c.ListObjectsResult, nil
//		}
//	}
//
//	return nil, errors.New("Can't find a response for the bucket and prefix combination")
//}

// TestProbeHandler iterates over a list of test cases
func TestProbeHandler(t *testing.T) {
	for _, c := range testCases {
		rr, err := probe(c.Bucket, c.Prefix)
		if err != nil {
			t.Errorf(err.Error())
		}

		c.testBody(rr.Body.String(), t)
	}
}

// ListObjectsV2 mocks out the corresponding function in the S3 client, returning the response that corresponds to the test case
//func (m *oss.Bucket) ListObjects(bucket, prefix string) (*oss.ListObjectsResult, error) {
//	r, err := testCases.response(bucket, prefix)
//	if err != nil {
//		return nil, err
//	}
//
//	return r, nil
//}

// Repeatable probe function
func probe(bucket, prefix string) (rr *httptest.ResponseRecorder, err error) {
	var uri string
	var client IClient
	//client = oss.Client{}
	if len(prefix) > 0 {
		uri = "/probe?bucket=" + bucket + "&prefix=" + prefix
	} else {
		uri = "/probe?bucket=" + bucket
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, client)
	})

	handler.ServeHTTP(rr, req)

	return
}
