package log

func callStructuredInfoForDepth(logger StructuredLogger) {
	logger.Info("structured callsite")
}

func callInstanceInfoForDepth(logger *gLogger) {
	logger.Info("instance callsite")
}

func callDefaultInfoForDepth() {
	infoDefault("default callsite")
}

func callInstanceVForDepth(logger *gLogger, level int) bool {
	return logger.V(level)
}

func callDefaultVForDepth(level int) bool {
	return vDefault(level)
}
