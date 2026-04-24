package dataclient

import (
	"time"

	"metricscapture/metricscapture/metrics"
)

func toAppTabularUploadArgs(req metrics.TabularUploadRequest) (
	tabularData []map[string]interface{},
	partID string,
	componentType string,
	componentName string,
	methodName string,
	dataRequestTimes [][2]time.Time,
	tags []string,
) {
	tabularData = make([]map[string]interface{}, 0, len(req.TabularData))
	for _, row := range req.TabularData {
		converted := make(map[string]interface{}, len(row))
		for k, v := range row {
			converted[k] = v
		}
		tabularData = append(tabularData, converted)
	}

	dataRequestTimes = make([][2]time.Time, 0, len(req.DataRequestTimes))
	for _, rt := range req.DataRequestTimes {
		dataRequestTimes = append(dataRequestTimes, [2]time.Time{rt.RequestedAt, rt.ReceivedAt})
	}

	return tabularData, req.PartID, req.ComponentType, req.ComponentName, req.MethodName, dataRequestTimes, req.Tags
}
