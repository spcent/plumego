package paging

// EndpointStatusParams selects a window of events and results for an
// endpoint status response.
type EndpointStatusParams struct {
	EventsPage      int
	EventsPageSize  int
	ResultsPage     int
	ResultsPageSize int
}

// NewEndpointStatusParams constructs an empty EndpointStatusParams.
func NewEndpointStatusParams() *EndpointStatusParams {
	return &EndpointStatusParams{}
}

// WithEvents sets the values for EventsPage and EventsPageSize.
func (p *EndpointStatusParams) WithEvents(page, pageSize int) *EndpointStatusParams {
	p.EventsPage = page
	p.EventsPageSize = pageSize
	return p
}

// WithResults sets the values for ResultsPage and ResultsPageSize.
func (p *EndpointStatusParams) WithResults(page, pageSize int) *EndpointStatusParams {
	p.ResultsPage = page
	p.ResultsPageSize = pageSize
	return p
}
