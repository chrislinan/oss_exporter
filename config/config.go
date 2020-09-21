package config

import "os"

var (
	AccessID   = os.Getenv("OSS_ACCESS_KEY_ID")
	AccessKey  = os.Getenv("OSS_ACCESS_KEY_SECRET")
	BucketName = os.Getenv("OSS_BUCKET")
	Endpoint   = os.Getenv("OSS_ENDPOINT")
)
