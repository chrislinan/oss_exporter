FROM quay.io/prometheus/busybox:latest

COPY oss_exporter /bin/oss_exporter

ENTRYPOINT ["/bin/oss_exporter"]