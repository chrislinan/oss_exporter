package config

import "os"

var (
	// Sample code's env configuration. You need to specify them with the actual configuration if you want to run sample code
	AccessID   = os.Getenv("OSS_ACCESS_KEY_ID")
	AccessKey  = os.Getenv("OSS_ACCESS_KEY_SECRET")
	BucketName = os.Getenv("OSS_BUCKET")
	Endpoint   = os.Getenv("OSS_ENDPOINT")

	// Credential
	//credentialAccessID  = os.Getenv("OSS_CREDENTIAL_KEY_ID")
	//credentialAccessKey = os.Getenv("OSS_CREDENTIAL_KEY_SECRET")
	//credentialUID       = os.Getenv("OSS_CREDENTIAL_UID")
	//
	//// The cname endpoint
	//endpoint4Cname = os.Getenv("OSS_CNAME_ENDPOINT")
)

//const (
//
//	// The object name in the sample code
//	objectKey       string = "my-object"
//	appendObjectKey string = "my-object-append"
//
//	// The local files to run sample code.
//	localFile          string = "sample/BingWallpaper-2015-11-07.jpg"
//	localCsvFile       string = "sample/sample_data.csv"
//	localJSONFile      string = "sample/sample_json.json"
//	localJSONLinesFile string = "sample/sample_json_lines.json"
//	htmlLocalFile      string = "sample/The Go Programming Language.html"
//)
