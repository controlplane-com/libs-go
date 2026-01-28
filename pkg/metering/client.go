package metering

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/controlplane-com/libs-go/pkg/checkpoints"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	"github.com/controlplane-com/libs-go/pkg/threading"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"github.com/controlplane-com/libs-go/pkg/web"
)

type Client interface {
	SetProfiling(enabled bool)
	QueryTagValues(token string, org string, query *TagValuesQueryRequest) (*TagValuesQueryResult, error)
	QueryByOrg(token string, org string, query *QueryRequest) (*QueryResult, error)
	Query(token string, query *QueryRequest) (*QueryResult, error)

	/*
	  DecomposeAndQueryByOrg and DecomposeAndQuery ensure that the given query is executed (at least) once for each
	  subSegment. If the subSegments are not aligned with query.TimeStep, the query will be decomposed into smaller
	  TimeSteps, given by decomposeInto.

	  Decomposition will fail if the query.TimeSegment is not aligned with the
	  smallest value in decomposeInto. e.g. if you give a query with a TimeSegment of
	  2023-01-01T00:01:01 - 2023-01-02T00:00:00, decomposition will fail because the start time is not aligned with the
	  smallest possible TimeStep (TimeStepHour)
	*/
	DecomposeAndQueryByOrg(token string, org string, query *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryResult, error)
	DecomposeAndQuery(token string, query *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryResult, error)
	GroupQueryResults(results []*QueryResult, aggregationPeriod timeUtils.TimeSegment, timeStep timeUtils.TimeStep, aggregateByTimeStep bool) (*QueryResult, error)

	AdminListCheckpoints(token string) (*checkpoints.ListCheckpointsResult, error)
	AdminResetCheckpoints(token string, request *checkpoints.ResetCheckpointsRequest) error
	AdminQueryCheckpoints(token string, request *checkpoints.QueryCheckpointsRequest) (*checkpoints.QueryCheckpointsResult, error)

	GetBucket(token string, request *GetBucketRequest) (*Bucket, error)
	ListBuckets(token string) ([]*Bucket, error)
}

func NewClient(url string) Client {
	return &httpClient{
		url:        url,
		httpClient: &http.Client{},
	}
}

type httpClient struct {
	httpClient       *http.Client
	url              string
	profilingEnabled bool
}

func (m *httpClient) SetProfiling(enabled bool) {
	m.profilingEnabled = enabled
}

func (m *httpClient) getFullUrl(path string) string {
	return fmt.Sprintf("%s%s", m.url, path)
}

func (m *httpClient) QueryTagValues(token string, org string, query *TagValuesQueryRequest) (*TagValuesQueryResult, error) {
	return web.DoJSONRequestWithBodyAndResult[*TagValuesQueryRequest, *TagValuesQueryResult](
		m.httpClient,
		http.MethodPost,
		m.getFullUrl(fmt.Sprintf("/org/%s/tags/values/query", org)),
		query,
		token)
}

func (m *httpClient) Query(token string, query *QueryRequest) (*QueryResult, error) {
	if err := ValidateQueryRequest(query); err != nil {
		return nil, err
	}
	return m.query(token, "/query", query)
}

func (m *httpClient) QueryByOrg(token string, org string, query *QueryRequest) (*QueryResult, error) {
	if err := ValidateQueryRequest(query); err != nil {
		return nil, err
	}
	return m.query(token, fmt.Sprintf("/org/%s/query", org), query)
}

func (m *httpClient) DecomposeAndQueryByOrg(token string, org string, query *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryResult, error) {
	if err := ValidateQueryRequest(query); err != nil {
		return nil, err
	}
	return m.decomposeAndQuery(token, fmt.Sprintf("/org/%s/query", org), query, boundaryTimes, decomposeInto)
}

func (m *httpClient) DecomposeAndQuery(token string, query *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryResult, error) {
	if err := ValidateQueryRequest(query); err != nil {
		return nil, err
	}
	return m.decomposeAndQuery(token, "/query", query, boundaryTimes, decomposeInto)
}

func (m *httpClient) GroupQueryResults(results []*QueryResult, aggregationPeriod timeUtils.TimeSegment, timeStep timeUtils.TimeStep, aggregateByTimeStep bool) (*QueryResult, error) {
	finalResults := &QueryResult{}
	aggregationMap := NewQueryResultAggregator(aggregationPeriod, timeStep, aggregateByTimeStep)
	err := aggregationMap.AggregateQueryResults(results...)
	if err != nil {
		return nil, err
	}

	finalResults, err = aggregationMap.QueryResult(timeStep)
	return finalResults, nil
}

func (m *httpClient) AdminListCheckpoints(token string) (*checkpoints.ListCheckpointsResult, error) {
	return web.DoJSONRequestWithResult[*checkpoints.ListCheckpointsResult](
		m.httpClient,
		http.MethodGet,
		m.getFullUrl("/checkpoints"),
		token,
	)
}

func (m *httpClient) AdminResetCheckpoints(token string, request *checkpoints.ResetCheckpointsRequest) error {
	return web.DoJSONRequestWithBody[*checkpoints.ResetCheckpointsRequest](
		m.httpClient,
		http.MethodPost,
		m.getFullUrl("/checkpoints/reset"),
		request,
		token,
	)
}

func (m *httpClient) AdminQueryCheckpoints(token string, request *checkpoints.QueryCheckpointsRequest) (*checkpoints.QueryCheckpointsResult, error) {
	return web.DoJSONRequestWithBodyAndResult[*checkpoints.QueryCheckpointsRequest, *checkpoints.QueryCheckpointsResult](
		m.httpClient,
		http.MethodPost,
		m.getFullUrl("/checkpoints/query"),
		request,
		token,
	)
}

func (m *httpClient) GetBucket(token string, request *GetBucketRequest) (*Bucket, error) {
	return web.DoJSONRequestWithResult[*Bucket](
		m.httpClient,
		http.MethodGet,
		m.getFullUrl(fmt.Sprintf("/bucket/%d", request.Id)),
		token,
	)
}

func (m *httpClient) ListBuckets(token string) ([]*Bucket, error) {
	return web.DoJSONRequestWithResult[[]*Bucket](
		m.httpClient,
		http.MethodGet,
		m.getFullUrl("/bucket"),
		token,
	)
}

func (m *httpClient) query(token string, urlPath string, query *QueryRequest) (*QueryResult, error) {
	defer func() {
		if !m.profilingEnabled {
			return
		}
		b, _ := json.Marshal(query)
		fmt.Println(string(b))
	}()
	return web.DoJSONRequestWithBodyAndResult[*QueryRequest, *QueryResult](
		m.httpClient,
		http.MethodPost,
		m.getFullUrl(urlPath),
		query,
		token,
	)
}

func (m *httpClient) decomposeAndQuery(token string, url string, request *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryResult, error) {
	queries, err := m.decomposeQuery(request, boundaryTimes, decomposeInto)
	if err != nil {
		return nil, err
	}
	results, err := threading.Parallel(queries, func(query *QueryRequest, _ int) ([]*QueryResult, error) {
		r, err := m.query(token, url, query)
		if err != nil {
			return nil, err
		}
		return []*QueryResult{r}, nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (m *httpClient) decomposeQuery(request *QueryRequest, boundaryTimes []time.Time, decomposeInto []timeUtils.TimeStep) ([]*QueryRequest, error) {
	segment := request.TimeSegment()
	decomposedSegments, err := segment.Decompose(boundaryTimes, decomposeInto, true)
	if err != nil {
		return nil, err
	}
	var queries []*QueryRequest
	for _, s := range decomposedSegments {
		cloned := request.Clone()
		if err := cloned.SetTimeSegment(s.TimeSegment); err != nil {
			return nil, err
		}
		cloned.TimeStep = s.TimeStep
		queries = append(queries, cloned)
	}
	return queries, nil
}

func ValidateQueryRequest(request *QueryRequest) error {
	if request == nil {
		return errors.New("nil query request given")
	}
	switch request.TimeStep {
	case timeUtils.TimeStepHour, timeUtils.TimeStepDay, timeUtils.TimeStepMonth, timeUtils.TimeStepWeek:
		break
	default:
		return errors.New(fmt.Sprintf("Invalid value for timeStep: %s. Valid values are: %s, %s, %s, and %s", request.TimeStep, timeUtils.TimeStepHour, timeUtils.TimeStepDay, timeUtils.TimeStepWeek, timeUtils.TimeStepMonth))
	}
	if !request.StartTime.Before(request.EndTime) {
		return errors.New(fmt.Sprintf("invalid values for startTime and endTime: (startTime=%s, endTime=%s). startTime must be before endTime", request.StartTime.Format(time.RFC3339), request.EndTime.Format(time.RFC3339)))
	}

	for _, groupKey := range request.GroupBy {
		for queryIndex, q := range request.ConsumptionQueries {
			if groupKey == GroupByConsumptionQuery {
				continue
			}
			if i := pipeline.IndexOf(q.AggregateBy, groupKey); i < 0 {
				return errors.New(fmt.Sprintf("the given metering query groups by %s, but consumption query %d does not aggregate by %s", groupKey, queryIndex, groupKey))
			}
		}
	}
	return nil
}

func boolToString(b bool) string {
	if b {
		return "does"
	}
	return "does not"
}
