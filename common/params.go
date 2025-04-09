package common

import (
	"encoding/json"
	"fmt"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"sync"
	"time"
)

// Param параметр
type Param struct {
	Name        string          `json:"name"`         // Имя параметра
	Value       float32         `json:"value"`        // Значение параметра
	LastUpdated *time.Time      `json:"last_updated"` // Время последнего обновления
	callbacks   []func(float32) // Список обратных вызовов
	mu          sync.RWMutex    // Мьютекс
}

func NewParam(name string) *Param {
	return &Param{Name: name}
}

func (p *Param) JSON() ([]byte, error) {
	return json.Marshal(p)
}

// Update обновить значения параметра
func (p *Param) Update(newValue float32) {
	p.mu.Lock()
	p.Value = newValue
	now := time.Now()
	p.LastUpdated = &now
	callbacks := make([]func(float32), len(p.callbacks))
	copy(callbacks, p.callbacks)
	p.mu.Unlock()

	for _, callback := range callbacks {
		go func(cb func(float32)) {
			cb(newValue)
		}(callback)
	}
}

func (p *Param) AddCallback(callback func(float32)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callbacks = append(p.callbacks, callback)
}

func (p *Param) RemoveCallback(callback func(float32)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cb := range p.callbacks {
		// Сравниваем указатели на функции
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			// Удаляем элемент
			p.callbacks = append(p.callbacks[:i], p.callbacks[i+1:]...)
			break
		}
	}
}

type ParamManager struct {
	params map[string]*Param
	mu     sync.RWMutex
}

func NewParamManager() *ParamManager {
	return &ParamManager{
		params: make(map[string]*Param),
	}
}

func (pm *ParamManager) Register(name string) *Param {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.params[name]; !exists {
		pm.params[name] = NewParam(name)
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
