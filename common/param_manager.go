package common

import (
	"fmt"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"sync"
	"time"
)

type ParamManager struct {
	params map[string]*Param
	mu     sync.RWMutex
}

func NewParamManager() *ParamManager {
	return &ParamManager{
		params: make(map[string]*Param),
	}
}

func (pm *ParamManager) Register(name string, options ...ParamOption) *Param {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.params[name]; !exists {
		pm.params[name] = NewParam(name, options...)
	}
	return pm.params[name]
}

func (pm *ParamManager) Get(name string) (*Param, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	param, exists := pm.params[name]
	return param, exists
}

func (pm *ParamManager) Update(name string, value float32) bool {
	pm.mu.RLock()
	param, exists := pm.params[name]
	pm.mu.RUnlock()

	if exists {
		fmt.Println("Update param:", name, value)
		param.Update(value)
		return true
	}
	return false
}

func (pm *ParamManager) RegisterCallback(name string, callback func(float32)) bool {
	pm.mu.RLock()
	param, exists := pm.params[name]
	pm.mu.RUnlock()

	if exists {
		param.AddCallback(callback)
		return true
	}
	return false
}

func (pm *ParamManager) HandleMessageParamValue(msg *ardupilotmega.MessageParamValue) {
	pm.Update(msg.ParamId, msg.ParamValue)
}

// GetParamsToUploadStartup список параметров загружаемых при старте
func (pm *ParamManager) GetParamsToUploadStartup() []*Param {
	var params []*Param
	for _, param := range pm.params {
		if param.UploadStartup && param.LastUpdated == nil {
			params = append(params, param)
		}
	}
	return params
}

// GetParamsToRefresh список параметров которые пора обновить
func (pm *ParamManager) GetParamsToRefresh() []*Param {
	var params []*Param
	for _, param := range pm.params {
		// Не пора ли обновить параметр
		if param.RefreshPeriod > 0 {
			if param.LastUpdated == nil || time.Now().Sub(*param.LastUpdated) > param.RefreshPeriod {
				params = append(params, param)
			}
		}
	}
	return params
}
