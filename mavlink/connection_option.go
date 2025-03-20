package mavlink

type ConnectionOption func(*Connection)

// WithParamManager Опция для установки ParamManager
func WithParamManager(pm *ParamManager) ConnectionOption {
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
