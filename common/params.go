package common

import (
	"encoding/json"
	"fmt"
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

	// Опции
	UploadStartup bool          // Загрузка значения при старте
	RefreshPeriod time.Duration // Период считывания
}

func NewParam(name string, options ...ParamOption) *Param {
	param := &Param{Name: name}

	// Применяем каждую опцию
	for _, option := range options {
		option(param)
	}

	return param
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
