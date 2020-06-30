package main

import (
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/chrislinan/oss_exporter/config"
)

const (
	namespace = "oss"
)

var (
	app           = kingpin.New(namespace+"_exporter", "Export metrics for oss certificates").DefaultEnvars()
	listenAddress = app.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9340").String()
	metricsPath   = app.Flag("web.metrics-path", "Path under which to expose metrics").Default("/metrics").String()
	probePath     = app.Flag("web.probe-path", "Path under which to expose the probe endpoint").Default("/probe").String()
	//endpoint      = app.Flag("oss.endpoint", "endpoint URL (required)").Default("").String()
	//bucketName    = app.Flag("oss.bucket-name", "Bucket name on alicloud OSS (required)").Default("").String()
)

var (
	ossListSuccess = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "list_success"),
		"If the ListObjects operation was a success",
		[]string{"bucket", "prefix"}, nil,
	)
	ossLastModifiedObjectDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_modified_object_date"),
		"The last modified date of the object that was modified most recently",
		[]string{"bucket", "prefix"}, nil,
	)
	ossLastModifiedObjectSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_modified_object_size_bytes"),
		"The size of the object that was modified most recently",
		[]string{"bucket", "prefix"}, nil,
	)
	ossObjectTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "objects_total"),
		"The total number of objects for the bucket/prefix combination",
		[]string{"bucket", "prefix"}, nil,
	)
	ossSumSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "objects_size_sum_bytes"),
		"The total size of all objects summed",
		[]string{"bucket", "prefix"}, nil,
	)
	ossBiggestSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "biggest_object_size_bytes"),
		"The size of the biggest object",
		[]string{"bucket", "prefix"}, nil,
	)
)

// Exporter is our exporter type
type Exporter struct {
	bucket string
	prefix string
	client IClient
}

type IClient interface {
	Bucket(name string) (IBucket, error)
}

type ClientWrapper struct {
	*oss.Client
}

func (cw ClientWrapper) Bucket(name string) (IBucket, error) {
	return cw.Client.Bucket(name)
}

type IBucket interface {
	ListObjects(options ...oss.Option) (oss.ListObjectsResult, error)
}

// Describe all the metrics we export
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- ossListSuccess
	ch <- ossLastModifiedObjectDate
	ch <- ossLastModifiedObjectSize
	ch <- ossObjectTotal
	ch <- ossSumSize
	ch <- ossBiggestSize
}

// Collect metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var lastModified time.Time
	var numberOfObjects float64
	var totalSize int64
	var biggestObjectSize int64
	var lastObjectSize int64

	bucket, err := e.client.Bucket(config.BucketName)
	if err != nil {
		log.Errorln(err)
		ch <- prometheus.MustNewConstMetric(
			ossListSuccess, prometheus.GaugeValue, 0, e.bucket, e.prefix,
		)
		return
	}

	//list all objects in bucket
	marker := ""
	for {
		lsRes, err := bucket.ListObjects(oss.Prefix(e.prefix), oss.Marker(marker))
		if err != nil {
			log.Errorln(err)
			ch <- prometheus.MustNewConstMetric(
				ossListSuccess, prometheus.GaugeValue, 0, e.bucket, e.prefix,
			)
			return
		}
		for _, item := range lsRes.Objects {
			numberOfObjects++
			totalSize = totalSize + item.Size
			if item.LastModified.After(lastModified) {
				lastModified = item.LastModified
				lastObjectSize = item.Size
			}
			if item.Size > biggestObjectSize {
				biggestObjectSize = item.Size
			}
		}
		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}

	ch <- prometheus.MustNewConstMetric(
		ossListSuccess, prometheus.GaugeValue, 1, e.bucket, e.prefix,
	)
	ch <- prometheus.MustNewConstMetric(
		ossLastModifiedObjectDate, prometheus.GaugeValue, float64(lastModified.UnixNano()/1e9), e.bucket, e.prefix,
	)
	ch <- prometheus.MustNewConstMetric(
		ossLastModifiedObjectSize, prometheus.GaugeValue, float64(lastObjectSize), e.bucket, e.prefix,
	)
	ch <- prometheus.MustNewConstMetric(
		ossObjectTotal, prometheus.GaugeValue, numberOfObjects, e.bucket, e.prefix,
	)
	ch <- prometheus.MustNewConstMetric(
		ossBiggestSize, prometheus.GaugeValue, float64(biggestObjectSize), e.bucket, e.prefix,
	)
	ch <- prometheus.MustNewConstMetric(
		ossSumSize, prometheus.GaugeValue, float64(totalSize), e.bucket, e.prefix,
	)
}

func probeHandler(w http.ResponseWriter, r *http.Request, client IClient) {

	bucket := r.URL.Query().Get("bucket")
	prefix := r.URL.Query().Get("prefix")

	exporter := &Exporter{
		bucket: bucket,
		prefix: prefix,
		client: client,
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	// Serve
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func init() {
	prometheus.MustRegister(version.NewCollector(namespace + "_exporter"))
}

func main() {

	log.AddFlags(kingpin.CommandLine)
	app.Version(version.Print(namespace + "_exporter"))
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// var sess *session.Session
	var err error

	if len(config.Endpoint) == 0 || len(config.BucketName) == 0 {
		log.Errorf("Please specify OSS_BUCKET and OSS_ENDPOINT environment variables")
		os.Exit(1)
	}

	if len(config.AccessKey) == 0 || len(config.AccessID) == 0 {
		log.Errorf("Please specify OSS_ACCESS_KEY_ID and OSS_ACCESS_KEY_SECRET environment variables")
		os.Exit(1)
	}

	client, err := oss.New(config.Endpoint, config.AccessID, config.AccessKey)
	c := ClientWrapper{client}
	if err != nil {
		log.Errorln("Error creating client ", err)
	}

	log.Infoln("Starting "+namespace+"_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc(*probePath, func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, c)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
						 <head><title>Alibaba OSS Exporter</title></head>
						 <body>
						 <h1>Alibaba OSS Exporter</h1>
						 <p><a href="` + *probePath + `?bucket=BUCKET&prefix=PREFIX">Query metrics for objects in BUCKET that match PREFIX</a></p>
						 <p><a href='` + *metricsPath + `'>Metrics</a></p>
						 </body>
						 </html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
