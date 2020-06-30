# AliCloud OSS Exporter
## This repo is forked from ribbybibby/s3_exporter and I made some changes so that this tool can be used in Alibaba Cloud.

This exporter provides metrics for OSS bucket objects by querying the API with a given bucket and prefix and constructing metrics based on the returned objects.

I find it useful for ensuring that backup jobs and batch uploads are functioning by comparing the growth in size/number of objects over time, or comparing the last modified date to an expected value.

## Building
```
make
```

## Running
Before runing this tool you need to set two environment variables: `OSS_ACCESS_KEY_ID` and `OSS_ACCESS_KEY_SECRET`
```
./oss_exporter <flags>
```

You can query a bucket and prefix combination by supplying them as parameters to /probe:
```
curl localhost:9340/probe?bucket=some-bucket&prefix=some-folder/some-file.txt
```


## Flags
    ./oss_exporter --help
 * __`--web.listen-address`:__ The port (default ":9340").
 * __`--web.metrics-path`:__ The path metrics are exposed under (default "/metrics")
 * __`--web.probe-path`:__ The path the probe endpoint is exposed under (default "/probe")
 * __`--oss.endpoint` :__ endpoint URL (required)
 * __`--oss.bucket-name` :__ Bucket name on alicloud OSS (required)
 * __`--version` :__ Show application version.

## Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| oss_biggest_object_size_bytes | The size of the largest object. | bucket, prefix |
| oss_last_modified_object_date | The modification date of the most recently modified object. | bucket, prefix |
| oss_last_modified_object_size_bytes | The size of the object that was modified most recently. | bucket, prefix |
| oss_list_success | Did the ListObjects operation complete successfully? | bucket, prefix |
| oss_objects_size_sum_bytes | The sum of the size of all the objects. | bucket, prefix |
| oss_objects_total | The total number of objects. | bucket, prefix |

## Prometheus
### Configuration
You should pass the params to a single instance of the exporter using relabelling, like so:
```yml
scrape_configs:
  - job_name: 'oss'
    metrics_path: /probe
    static_configs:
      - targets:
        - bucket=stuff;prefix=thing.txt;
        - bucket=other-stuff;prefix=another-thing.gif;
    relabel_configs:
      - source_labels: [__address__]
        regex: '^bucket=(.*);prefix=(.*);$'
        replacement: '${1}'
        target_label: '__param_bucket'
      - source_labels: [__address__]
        regex: '^bucket=(.*);prefix=(.*);$'
        replacement: '${2}'
        target_label: '__param_prefix'
      - target_label: __address__
        replacement: 127.0.0.1:9340  # oss exporter.

```
### Example Queries
Return series where the last modified object date is more than 24 hours ago:
```
(time() - s3_last_modified_object_date) / 3600 > 24
```
