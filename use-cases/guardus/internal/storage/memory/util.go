package memory

import (
	"guardus/internal/domain/endpoint"
	"guardus/internal/storage/common/paging"
)

// ShallowCopyEndpointStatus returns a paginated shallow copy of ss.
func ShallowCopyEndpointStatus(ss *endpoint.Status, params *paging.EndpointStatusParams) *endpoint.Status {
	out := &endpoint.Status{
		Name:   ss.Name,
		Group:  ss.Group,
		Key:    ss.Key,
		Uptime: endpoint.NewUptime(),
	}
	if params == nil || (params.ResultsPage == 0 && params.ResultsPageSize == 0 && params.EventsPage == 0 && params.EventsPageSize == 0) {
		out.Results = ss.Results
		out.Events = ss.Events
		return out
	}
	resultsStart, resultsEnd := getStartAndEndIndex(len(ss.Results), params.ResultsPage, params.ResultsPageSize)
	if resultsStart < 0 || resultsEnd < 0 {
		out.Results = []*endpoint.Result{}
	} else {
		out.Results = ss.Results[resultsStart:resultsEnd]
	}
	eventsStart, eventsEnd := getStartAndEndIndex(len(ss.Events), params.EventsPage, params.EventsPageSize)
	if eventsStart < 0 || eventsEnd < 0 {
		out.Events = []*endpoint.Event{}
	} else {
		out.Events = ss.Events[eventsStart:eventsEnd]
	}
	return out
}

func getStartAndEndIndex(numberOfResults, page, pageSize int) (int, int) {
	if page < 1 || pageSize < 0 {
		return -1, -1
	}
	start := numberOfResults - (page * pageSize)
	end := numberOfResults - ((page - 1) * pageSize)
	if start > numberOfResults {
		start = -1
	} else if start < 0 {
		start = 0
	}
	if end > numberOfResults {
		end = numberOfResults
	}
	return start, end
}

// AddResult appends result to ss.Results, emits events on success/failure
// transitions, and trims slices to the configured maxima.
func AddResult(ss *endpoint.Status, result *endpoint.Result, maximumNumberOfResults, maximumNumberOfEvents int) {
	if ss == nil {
		return
	}
	if len(ss.Results) > 0 {
		if ss.Results[len(ss.Results)-1].Success != result.Success {
			ss.Events = append(ss.Events, endpoint.NewEventFromResult(result))
			if len(ss.Events) > maximumNumberOfEvents {
				ss.Events = ss.Events[len(ss.Events)-maximumNumberOfEvents:]
			}
		}
	} else {
		ss.Events = append(ss.Events, endpoint.NewEventFromResult(result))
	}
	ss.Results = append(ss.Results, result)
	if len(ss.Results) > maximumNumberOfResults {
		ss.Results = ss.Results[len(ss.Results)-maximumNumberOfResults:]
	}
	processUptimeAfterResult(ss.Uptime, result)
}
