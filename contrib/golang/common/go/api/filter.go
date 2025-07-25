/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import "google.golang.org/protobuf/types/known/anypb"

type (
	// PassThroughStreamEncoderFilter provides the no-op implementation of the StreamEncoderFilter interface.
	PassThroughStreamEncoderFilter struct{}
	// PassThroughStreamDecoderFilter provides the no-op implementation of the StreamDecoderFilter interface.
	PassThroughStreamDecoderFilter struct{}
	// PassThroughStreamFilter provides the no-op implementation of the StreamFilter interface.
	PassThroughStreamFilter struct {
		PassThroughStreamDecoderFilter
		PassThroughStreamEncoderFilter
	}

	// EmptyDownstreamFilter provides the no-op implementation of the DownstreamFilter interface
	EmptyDownstreamFilter struct{}
	// EmptyUpstreamFilter provides the no-op implementation of the UpstreamFilter interface
	EmptyUpstreamFilter struct{}

	// PassThroughHttpTcpBridge provides the no-op implementation of the HttpTcpBridge interface
	PassThroughHttpTcpBridge struct{}
)

// request
type StreamDecoderFilter interface {
	DecodeHeaders(RequestHeaderMap, bool) StatusType
	DecodeData(BufferInstance, bool) StatusType
	DecodeTrailers(RequestTrailerMap) StatusType
}

func (*PassThroughStreamDecoderFilter) DecodeHeaders(RequestHeaderMap, bool) StatusType {
	return Continue
}

func (*PassThroughStreamDecoderFilter) DecodeData(BufferInstance, bool) StatusType {
	return Continue
}

func (*PassThroughStreamDecoderFilter) DecodeTrailers(RequestTrailerMap) StatusType {
	return Continue
}

// response
type StreamEncoderFilter interface {
	EncodeHeaders(ResponseHeaderMap, bool) StatusType
	EncodeData(BufferInstance, bool) StatusType
	EncodeTrailers(ResponseTrailerMap) StatusType
}

func (*PassThroughStreamEncoderFilter) EncodeHeaders(ResponseHeaderMap, bool) StatusType {
	return Continue
}

func (*PassThroughStreamEncoderFilter) EncodeData(BufferInstance, bool) StatusType {
	return Continue
}

func (*PassThroughStreamEncoderFilter) EncodeTrailers(ResponseTrailerMap) StatusType {
	return Continue
}

type StreamFilter interface {
	// http request
	StreamDecoderFilter
	// response stream
	StreamEncoderFilter

	// log
	OnLog(RequestHeaderMap, RequestTrailerMap, ResponseHeaderMap, ResponseTrailerMap)
	OnLogDownstreamStart(RequestHeaderMap)
	OnLogDownstreamPeriodic(RequestHeaderMap, RequestTrailerMap, ResponseHeaderMap, ResponseTrailerMap)

	// destroy filter
	OnDestroy(DestroyReason)
	OnStreamComplete()
}

func (*PassThroughStreamFilter) OnLog(RequestHeaderMap, RequestTrailerMap, ResponseHeaderMap, ResponseTrailerMap) {
}

func (*PassThroughStreamFilter) OnLogDownstreamStart(RequestHeaderMap) {
}

func (*PassThroughStreamFilter) OnLogDownstreamPeriodic(RequestHeaderMap, RequestTrailerMap, ResponseHeaderMap, ResponseTrailerMap) {
}

func (*PassThroughStreamFilter) OnDestroy(DestroyReason) {
}

func (*PassThroughStreamFilter) OnStreamComplete() {
}

type Config interface {
	// Called when the current config is deleted due to an update or removal of plugin.
	// You can use this method is you store some resources in the config to be released later.
	Destroy()
}

type StreamFilterConfigParser interface {
	// Parse the proto message to any Go value, and return error to reject the config.
	// This is called when Envoy receives the config from the control plane.
	// Also, you can define Metrics through the callbacks, and the callbacks will be nil when parsing the route config.
	// You can return a config implementing the Config interface if you need fine control over its lifecycle.
	Parse(any *anypb.Any, callbacks ConfigCallbackHandler) (interface{}, error)
	// Merge the two configs(filter level config or route level config) into one.
	// May merge multi-level configurations, i.e. filter level, virtualhost level, router level and weighted cluster level,
	// into a single one recursively, by invoking this method multiple times.
	// You can return a config implementing the Config interface if you need fine control over its lifecycle.
	Merge(parentConfig interface{}, childConfig interface{}) interface{}
}

type StreamFilterFactory func(config interface{}, callbacks FilterCallbackHandler) StreamFilter

// stream info
// refer https://github.com/envoyproxy/envoy/blob/main/envoy/stream_info/stream_info.h
type StreamInfo interface {
	GetRouteName() string
	FilterChainName() string
	// Protocol return the request's protocol.
	Protocol() (string, bool)
	// ResponseCode return the response code.
	ResponseCode() (uint32, bool)
	// ResponseCodeDetails return the response code details.
	ResponseCodeDetails() (string, bool)
	// AttemptCount return the number of times the request was attempted upstream.
	AttemptCount() uint32
	// Get the dynamic metadata of the request
	DynamicMetadata() DynamicMetadata
	// DownstreamLocalAddress return the downstream local address.
	DownstreamLocalAddress() string
	// DownstreamRemoteAddress return the downstream remote address.
	DownstreamRemoteAddress() string
	// UpstreamLocalAddress return the upstream local address.
	UpstreamLocalAddress() (string, bool)
	// UpstreamRemoteAddress return the upstream remote address.
	UpstreamRemoteAddress() (string, bool)
	// UpstreamClusterName return the upstream host cluster.
	UpstreamClusterName() (string, bool)
	// FilterState return the filter state interface.
	FilterState() FilterState
	// VirtualClusterName returns the name of the virtual cluster which got matched
	VirtualClusterName() (string, bool)
	// WorkerID returns the ID of the Envoy worker thread
	WorkerID() uint32
	// Some fields in stream info can be fetched via GetProperty
	// For example, startTime() is equal to GetProperty("request.time")
}

type StreamFilterCallbacks interface {
	StreamInfo() StreamInfo

	// ClearRouteCache clears the route cache for the current request, and filtermanager will re-fetch the route in the next filter.
	// Please be careful to invoke it, since filtermanager will raise an 404 route_not_found response when failed to re-fetch a route.
	ClearRouteCache()
	// RefreshRouteCache works like ClearRouteCache, but it will re-fetch the route immediately.
	RefreshRouteCache()
	Log(level LogType, msg string)
	LogLevel() LogType
	// GetProperty fetch Envoy attribute and return the value as a string.
	// The list of attributes can be found in https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes.
	// If the fetch succeeded, a string will be returned.
	// If the value is a timestamp, it is returned as a timestamp string like "2023-07-31T07:21:40.695646+00:00".
	// If the fetch failed (including the value is not found), an error will be returned.
	//
	// The error can be one of:
	// * ErrInternalFailure
	// * ErrSerializationFailure (Currently, fetching attributes in List/Map type are unsupported)
	// * ErrValueNotFound
	GetProperty(key string) (string, error)
	// TODO add more for filter callbacks

	// Get secret manager.
	// Secrets should be defined in the plugin configuration.
	// It is safe to use this secret manager from any goroutine.
	SecretManager() SecretManager
}

// FilterProcessCallbacks is the interface for filter to process request/response in decode/encode phase.
type FilterProcessCallbacks interface {
	// Continue or SendLocalReply should be last API invoked, no more code after them.
	Continue(StatusType)
	SendLocalReply(responseCode int, bodyText string, headers map[string][]string, grpcStatus int64, details string)
	// RecoverPanic recover panic in defer and terminate the request by SendLocalReply with 500 status code.
	RecoverPanic()
	// AddData add extra data when processing headers/trailers.
	// For example, turn a headers only request into a request with a body, add more body when processing trailers, and so on.
	// The second argument isStreaming supplies if this caller streams data or buffers the full body.
	AddData(data []byte, isStreaming bool)
	// InjectData inject the content of slice data via Envoy StreamXXFilterCallbacks's injectXXDataToFilterChaininjectData.
	InjectData(data []byte)
}

type DecoderFilterCallbacks interface {
	FilterProcessCallbacks
	// Sets an upstream address override for the request. When the overridden host exists in the host list of the routed cluster
	// and can be selected directly, the load balancer bypasses its algorithm and routes traffic directly to the specified host.
	//
	// Here are some cases:
	// 1. Set a valid host(no matter in or not in the cluster), will route to the specified host directly and return 200.
	// 2. Set a non-IP host, C++ side will return error and not route to cluster.
	// 3. Set a unavaiable host, and the host is not in the cluster, will req the valid host in the cluster and rerurn 200.
	// 4. Set a unavaiable host, and the host is in the cluster, but not available(can not connect to the host), will req the unavaiable hoat and rerurn 503.
	// 5. Set a unavaiable host, and the host is in the cluster, but not available(can not connect to the host), and with retry. when first request with unavaiable host failed 503, the second request will retry with the valid host, then the second request will succeed and finally return 200.
	// 6. Set a unavaiable host with strict mode, and the host is in the cluster, will req the unavaiable host and rerurn 503.
	// 7. Set a unavaiable host with strict mode, and the host is not in the cluster, will req the unavaiable host and rerurn 503.
	// 8. Set a unavaiable host with strict mode and retry. when first request with unavaiable host failed 503, the second request will retry with the valid host, then the second request will succeed and finally return 200.
	// 9. Set a unavaiable host with strict mode and retry, and the host is not in the cluster, will req the unavaiable host and rerurn 503.
	//
	// The function takes two arguments:
	//
	// host (string): The upstream host address to use for the request. This must be a valid IP address(with port); otherwise, the
	// C++ side will throw an error.
	//
	// strict (boolean): Determines whether the HTTP request must be strictly routed to the requested
	// host. When set to ``true``, if the requested host is invalid, Envoy will return a 503 status code.
	// The default value is ``false``, which allows Envoy to fall back to its load balancing mechanism. In this case, if the
	// requested host is invalid, the request will be routed according to the load balancing algorithm and choose other hosts.
	SetUpstreamOverrideHost(host string, strict bool) error
}

type EncoderFilterCallbacks interface {
	FilterProcessCallbacks
}

type FilterCallbackHandler interface {
	StreamFilterCallbacks
	// DecoderFilterCallbacks could only be used in DecodeXXX phases.
	DecoderFilterCallbacks() DecoderFilterCallbacks
	// EncoderFilterCallbacks could only be used in EncodeXXX phases.
	EncoderFilterCallbacks() EncoderFilterCallbacks
}

type DynamicMetadata interface {
	Get(filterName string) map[string]interface{}
	Set(filterName string, key string, value interface{})
}

type DownstreamFilter interface {
	// Called when a connection is first established.
	OnNewConnection() FilterStatus
	// Called when data is read on the connection.
	OnData(buffer []byte, endOfStream bool) FilterStatus
	// Callback for connection events.
	OnEvent(event ConnectionEvent)
	// Called when data is to be written on the connection.
	OnWrite(buffer []byte, endOfStream bool) FilterStatus
}

func (*EmptyDownstreamFilter) OnNewConnection() FilterStatus {
	return NetworkFilterContinue
}

func (*EmptyDownstreamFilter) OnData(buffer []byte, endOfStream bool) FilterStatus {
	return NetworkFilterContinue
}

func (*EmptyDownstreamFilter) OnEvent(event ConnectionEvent) {
}

func (*EmptyDownstreamFilter) OnWrite(buffer []byte, endOfStream bool) FilterStatus {
	return NetworkFilterContinue
}

type UpstreamFilter interface {
	// Called when a connection is available to process a request/response.
	OnPoolReady(cb ConnectionCallback)
	// Called when a pool error occurred and no connection could be acquired for making the request.
	OnPoolFailure(poolFailureReason PoolFailureReason, transportFailureReason string)
	// Invoked when data is delivered from the upstream connection.
	OnData(buffer []byte, endOfStream bool)
	// Callback for connection events.
	OnEvent(event ConnectionEvent)
}

func (*EmptyUpstreamFilter) OnPoolReady(cb ConnectionCallback) {
}

func (*EmptyUpstreamFilter) OnPoolFailure(poolFailureReason PoolFailureReason, transportFailureReason string) {
}

func (*EmptyUpstreamFilter) OnData(buffer []byte, endOfStream bool) FilterStatus {
	return NetworkFilterContinue
}

func (*EmptyUpstreamFilter) OnEvent(event ConnectionEvent) {
}

type ConnectionCallback interface {
	// StreamInfo returns the stream info of the connection
	StreamInfo() StreamInfo
	// Write data to the connection.
	Write(buffer []byte, endStream bool)
	// Close the connection.
	Close(closeType ConnectionCloseType)
	// EnableHalfClose only for upstream connection
	EnableHalfClose(enabled bool)
}

type StateType int

const (
	StateTypeReadOnly StateType = 0
	StateTypeMutable  StateType = 1
)

type LifeSpan int

const (
	LifeSpanFilterChain LifeSpan = 0
	LifeSpanRequest     LifeSpan = 1
	LifeSpanConnection  LifeSpan = 2
	LifeSpanTopSpan     LifeSpan = 3
)

type StreamSharing int

const (
	None                             StreamSharing = 0
	SharedWithUpstreamConnection     StreamSharing = 1
	SharedWithUpstreamConnectionOnce StreamSharing = 2
)

type FilterState interface {
	SetString(key, value string, stateType StateType, lifeSpan LifeSpan, streamSharing StreamSharing)
	GetString(key string) string
}

type SecretManager interface {
	// Get generic secret from secret manager.
	// bool is false on missing secret
	GetGenericSecret(name string) (string, bool)
}

type MetricType uint32

const (
	Counter   MetricType = 0
	Gauge     MetricType = 1
	Histogram MetricType = 2
)

type ConfigCallbacks interface {
	// Define a metric, for different MetricType, name must be different,
	// for same MetricType, the same name will share a metric.
	DefineCounterMetric(name string) CounterMetric
	DefineGaugeMetric(name string) GaugeMetric
	// TODO Histogram
}

type ConfigCallbackHandler interface {
	ConfigCallbacks
}

type CounterMetric interface {
	Increment(offset int64)
	Get() uint64
	Record(value uint64)
}

type GaugeMetric interface {
	Increment(offset int64)
	Get() uint64
	Record(value uint64)
}

// TODO
type HistogramMetric interface {
}

type HttpTcpBridgeCallbackHandler interface {
	// GetRouteName returns the name of the route which got matched
	GetRouteName() string
	// GetVirtualClusterName returns the name of the virtual cluster which got matched
	GetVirtualClusterName() string
	// SetSelfHalfCloseForUpstreamConn default is false
	SetSelfHalfCloseForUpstreamConn(enabled bool)
}

type HttpTcpBridge interface {

	// Invoked when header is delivered from the downstream.
	// Notice-1: when return HttpTcpBridgeContinue or HttpTcpBridgeStopAndBuffer, dataForSet is used to be sent to upstream; when return HttpTcpBridgeEndStream, dataForSet is useed to sent to downstream as response body.
	// Notice-2: headerMap and dataToUpstream cannot be invoked after the func return.
	EncodeHeaders(headerMap RequestHeaderMap, dataForSet BufferInstance, endOfStream bool) HttpTcpBridgeStatus

	// Streaming, Invoked when data is delivered from the downstream.
	// Notice: buffer cannot be invoked after the func return.
	EncodeData(buffer BufferInstance, endOfStream bool) HttpTcpBridgeStatus

	// Streaming, Called when data is read on from tcp upstream.
	// Notice-1: when return HttpTcpBridgeContinue, resp headers will be send to http all at once; from then on, you MUST NOT invoke responseHeaderForSet at any time(or you will get panic).
	// Notice-2: responseHeaderForSet and buffer cannot be invoked after the func return.
	OnUpstreamData(responseHeaderForSet ResponseHeaderMap, buffer BufferInstance, endOfStream bool) HttpTcpBridgeStatus

	// destroy filter
	OnDestroy()
}

func (*PassThroughHttpTcpBridge) EncodeHeaders(headerMap RequestHeaderMap, dataForSet BufferInstance, endOfStream bool) HttpTcpBridgeStatus {
	return HttpTcpBridgeContinue
}

func (*PassThroughHttpTcpBridge) EncodeData(buffer BufferInstance, endOfStream bool) HttpTcpBridgeStatus {
	return HttpTcpBridgeContinue
}

func (*PassThroughHttpTcpBridge) OnUpstreamData(responseHeaderForSet ResponseHeaderMap, buffer BufferInstance, endOfStream bool) HttpTcpBridgeStatus {
	return HttpTcpBridgeContinue
}

func (*PassThroughHttpTcpBridge) OnDestroy() {
}

type HttpTcpBridgeFactory func(config interface{}, callbacks HttpTcpBridgeCallbackHandler) HttpTcpBridge

type HttpTcpBridgeConfigParser interface {
	Parse(any *anypb.Any) (interface{}, error)
}
