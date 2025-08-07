package common

import "time"

type ParamOption func(*Param)

// WithUploadStartup Считать параметр при старте
func WithUploadStartup(uploadStartup bool) ParamOption {
	return func(param *Param) {
		param.UploadStartup = uploadStartup
	}
}

// WithRefreshPeriod Период считывания значения, если очень критично
func WithRefreshPeriod(refreshPeriodSec time.Duration) ParamOption {
	return func(param *Param) {
		param.RefreshPeriodSec = refreshPeriodSec
	}
}
