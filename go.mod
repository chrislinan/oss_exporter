module github.com/chrislinan/oss_exporter

require (
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/aliyun/aliyun-oss-go-sdk v2.1.2+incompatible
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/jarcoal/httpmock v1.0.5
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.10.0
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace (
	golang.org/x/sys v0.0.0-20200513112337-417ce2331b5c => github.com/golang/sys v0.0.0-20200513112337-417ce2331b5c
	golang.org/x/text v0.3.2 => github.com/golang/text v0.3.2
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 => github.com/golang/time v0.0.0-20200416051211-89c76fbcd5d1
)

go 1.12
