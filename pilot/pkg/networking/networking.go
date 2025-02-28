// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"istio.io/istio/pkg/config/protocol"
)

// ListenerProtocol is the protocol associated with the listener.
type ListenerProtocol int

const (
	// ListenerProtocolUnknown is an unknown type of listener.
	ListenerProtocolUnknown = iota
	// ListenerProtocolTCP is a TCP listener.
	ListenerProtocolTCP
	// ListenerProtocolHTTP is an HTTP listener.
	ListenerProtocolHTTP
	// ListenerProtocolAuto enables auto protocol detection
	ListenerProtocolAuto
)

// ModelProtocolToListenerProtocol converts from a config.Protocol to its corresponding plugin.ListenerProtocol
func ModelProtocolToListenerProtocol(p protocol.Instance) ListenerProtocol {
	switch p {
	case protocol.HTTP, protocol.HTTP2, protocol.HTTP_PROXY, protocol.GRPC, protocol.GRPCWeb:
		return ListenerProtocolHTTP
	case protocol.TCP, protocol.HTTPS, protocol.TLS,
		protocol.Mongo, protocol.Redis, protocol.MySQL:
		return ListenerProtocolTCP
	case protocol.UDP:
		return ListenerProtocolUnknown
	case protocol.Unsupported:
		return ListenerProtocolAuto
	default:
		// Should not reach here.
		return ListenerProtocolAuto
	}
}

type TransportProtocol uint8

const (
	// TransportProtocolTCP is a TCP listener
	TransportProtocolTCP = iota
	// TransportProtocolQUIC is a QUIC listener
	TransportProtocolQUIC
)

func (tp TransportProtocol) String() string {
	switch tp {
	case TransportProtocolTCP:
		return "tcp"
	case TransportProtocolQUIC:
		return "quic"
	}
	return "unknown"
}

func (tp TransportProtocol) ToEnvoySocketProtocol() core.SocketAddress_Protocol {
	if tp == TransportProtocolQUIC {
		return core.SocketAddress_UDP
	}
	return core.SocketAddress_TCP
}

// FilterChain describes a set of filters (HTTP or TCP) with a shared TLS context.
type FilterChain struct {
	// FilterChainMatch is the match used to select the filter chain.
	FilterChainMatch *listener.FilterChainMatch
	// TLSContext is the TLS settings for this filter chains.
	TLSContext *tls.DownstreamTlsContext
	// ListenerProtocol indicates whether this filter chain is for HTTP or TCP
	// Note that HTTP filter chains can also have network filters
	ListenerProtocol ListenerProtocol
	// TransportProtocol indicates the type of transport used - TCP, UDP, QUIC
	// This would be TCP by default
	TransportProtocol TransportProtocol

	// HTTP is the set of HTTP filters for this filter chain
	HTTP []*hcm.HttpFilter
	// TCP is the set of network (TCP) filters for this filter chain.
	TCP []*listener.Filter
}

// MutableObjects is a set of objects passed to On*Listener callbacks. Fields may be nil or empty.
// Any lists should not be overridden, but rather only appended to.
// Non-list fields may be mutated; however it's not recommended to do this since it can affect other plugins in the
// chain in unpredictable ways.
// TODO: do we need this now?
type MutableObjects struct {
	// Listener is the listener being built. Must be initialized before Plugin methods are called.
	Listener *listener.Listener

	// FilterChains is the set of filter chains that will be attached to Listener.
	FilterChains []FilterChain
}

// ListenerClass defines the class of the listener
type ListenerClass int

const (
	ListenerClassUndefined ListenerClass = iota
	ListenerClassSidecarInbound
	ListenerClassSidecarOutbound
	ListenerClassGateway
)
