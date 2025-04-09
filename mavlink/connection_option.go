package mavlink

import "gomavclient/common"

type ConnectionOption func(*Connection)

// WithParamManager Опция для работы с параметрами
func WithParamManager(pm *common.ParamManager) ConnectionOption {
	return func(c *Connection) {
		c.paramManager = pm
	}
}

// WithDebug Опция для включения/выключения отладки
func WithDebug(debug bool) ConnectionOption {
	return func(c *Connection) {
		c.debug = debug
	}
}

// WithTelemetryManager Опция для работы с телеметрией
func WithTelemetryManager(tm *TelemetryManager) ConnectionOption {
	return func(c *Connection) {
		c.TelemetryManager = tm
	}
}
