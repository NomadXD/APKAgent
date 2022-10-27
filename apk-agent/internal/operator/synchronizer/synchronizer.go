package synchronizer

import "github.com/AmaliMatharaarachchi/APKAgent/apk-agent/internal/logger"

func ListenToEvents(apiStore *APIStore, apiMappingStore *APIToHttpRoutesMapping, ch *chan string) {
	for event := range *ch {
		logger.LoggerOperator.Infof("Event received: %v\n", event)
	}
}
