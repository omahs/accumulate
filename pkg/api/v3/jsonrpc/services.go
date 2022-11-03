package jsonrpc

import (
	"context"
	"encoding/json"

	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
)

const ErrCodeProtocol = -33000

func parseRequest[T any](input json.RawMessage) (T, error) {
	var v T
	if len(input) == 0 {
		return v, errors.BadRequest.With("empty request")
	}
	err := json.Unmarshal(input, &v)
	if err != nil {
		return v, errors.EncodingError.WithFormat("unmarshal request: %w", err)
	}
	return v, nil
}

func formatResponse(res interface{}, err error) interface{} {
	if err == nil {
		return res
	}

	// Ensure the error is an Error
	err2 := errors.UnknownError.Wrap(err).(*errors.Error)
	return jsonrpc2.NewError(ErrCodeProtocol-jsonrpc2.ErrorCode(err2.Code), err2.Code.String(), err2)
}

type NodeService struct{ api.NodeService }

func (s NodeService) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"node-status": s.NodeStatus,
	}
}

func (s NodeService) NodeStatus(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*NodeStatusRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.NodeService.NodeStatus(ctx, req.NodeStatusOptions))
}

type NetworkService struct{ api.NetworkService }

func (s NetworkService) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"network-status": s.NetworkStatus,
	}
}

func (s NetworkService) NetworkStatus(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*NetworkStatusRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.NetworkService.NetworkStatus(ctx, req.NetworkStatusOptions))
}

type MetricsService struct{ api.MetricsService }

func (s MetricsService) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"metrics": s.Metrics,
	}
}

func (s MetricsService) Metrics(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*MetricsRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.MetricsService.Metrics(ctx, req.MetricsOptions))
}

type Querier struct{ api.Querier }

func (s Querier) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"query": s.Query,
	}
}

func (s Querier) Query(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*QueryRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.Querier.Query(ctx, req.Scope, req.Query))
}

type Submitter struct{ api.Submitter }

func (s Submitter) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"submit": s.Submit,
	}
}

func (s Submitter) Submit(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*SubmitRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.Submitter.Submit(ctx, req.Envelope, req.SubmitOptions))
}

type Validator struct{ api.Validator }

func (s Validator) methods() jsonrpc2.MethodMap {
	return jsonrpc2.MethodMap{
		"validate": s.Validate,
	}
}

func (s Validator) Validate(ctx context.Context, params json.RawMessage) interface{} {
	req, err := parseRequest[*ValidateRequest](params)
	if err != nil {
		return formatResponse(nil, err)
	}
	return formatResponse(s.Validator.Validate(ctx, req.Envelope, req.ValidateOptions))
}
