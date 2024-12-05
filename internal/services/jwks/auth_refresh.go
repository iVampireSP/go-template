package jwks

import "time"

var refreshRate = 1 * time.Hour

func (j *JWKS) SetupAuthRefresh() {
	// 先刷新一次
	j.RefreshJWKS()
	var firstRefreshed = true

	// 启动一个定时器
	go func() {
		for {
			if firstRefreshed {
				firstRefreshed = false
			} else {
				j.RefreshJWKS()
			}
			time.Sleep(refreshRate)
		}
	}()
}
