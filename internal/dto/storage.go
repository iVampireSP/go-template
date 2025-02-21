package dto

type StorageCredentials struct {
	Name      string `json:"name"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	Endpoint  string `json:"endpoint"`
	Prefix    string `json:"prefix"`
}

type ChunkDownloadURL struct {
	Hash  string `json:"hash"`
	Urls  []Url  `json:"urls"`
	Order uint64 `json:"order"`
}

type Url struct {
	StorageName string `json:"storage_name"`
	Url         string `json:"url"`
}

// func (s *StorageCredentials) GetClient() (*minio.Client, error) {
// 	// 如果是 https 则使用 ssl
// 	var useSSL bool
// 	if strings.HasPrefix(s.Endpoint, "https") {
// 		useSSL = true
// 	} else {
// 		useSSL = false
// 	}

// 	// 去除协议
// 	endpoint := strings.TrimPrefix(s.Endpoint, "http://")
// 	endpoint = strings.TrimPrefix(endpoint, "https://")

// 	return minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(s.AccessKey, s.SecretKey, ""),
// 		Secure: useSSL,
// 		Region: s.Region,
// 	})
// }
