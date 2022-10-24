package api

import (
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type ValueOrError[T any] struct {
	Value T
	Err   error
}

type BatchedNodeService interface {
	NodeStatus(opts NodeStatusOptions) <-chan ValueOrError[*NodeStatus]
}

type BatchedNetworkService interface {
	NetworkStatus(opts NetworkStatusOptions) <-chan ValueOrError[*NetworkStatus]
}

type BatchedMetricsService interface {
	Metrics(opts MetricsOptions) <-chan ValueOrError[*Metrics]
}

type BatchedQueryService interface {
	Query(scope *url.URL, query Query) <-chan ValueOrError[Record]
}

type BatchedSubmitService interface {
	Submit(envelope *protocol.Envelope, opts SubmitOptions) <-chan ValueOrError[[]*Submission]
}

type BatchedValidateService interface {
	Validate(envelope *protocol.Envelope, opts ValidateOptions) <-chan ValueOrError[[]*Submission]
}
