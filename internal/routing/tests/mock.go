// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License, Version 2.0.
package tests

import (
	"context"
	"sync"

	"github.com/azure/peerd/internal/routing"
	"github.com/azure/peerd/pkg/mocks"
	"github.com/azure/peerd/pkg/peernet"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MockRouter struct {
	net      peernet.Network
	mx       sync.RWMutex
	resolver map[string][]string

	negCache map[string]struct{}
}

// Net implements routing.Router.
func (m *MockRouter) Net() peernet.Network {
	return m.net
}

// ResolveWithCache implements Router.
func (m *MockRouter) ResolveWithCache(ctx context.Context, key string, allowSelf bool, count int) (<-chan routing.PeerInfo, func(), error) {
	c, e := m.Resolve(ctx, key, allowSelf, count)
	return c, func() {
		m.mx.Lock()
		defer m.mx.Unlock()
		m.negCache[key] = struct{}{}
	}, e
}

var _ routing.Router = &MockRouter{}

func NewMockRouter(resolver map[string][]string) *MockRouter {
	n, err := peernet.New(&mocks.MockHost{PeerStore: &mocks.MockPeerstore{}})
	if err != nil {
		panic(err)
	}

	return &MockRouter{
		net:      n,
		resolver: resolver,
		negCache: map[string]struct{}{},
	}
}

func (m *MockRouter) Close() error {
	return nil
}

func (m *MockRouter) Resolve(ctx context.Context, key string, allowSelf bool, count int) (<-chan routing.PeerInfo, error) {
	peerCh := make(chan routing.PeerInfo, count)
	peers, ok := m.resolver[key]
	// Not found will look forever until timeout.
	if !ok {
		return peerCh, nil
	}

	go func() {
		m.mx.RLock()
		defer m.mx.RUnlock()
		for _, p := range peers {
			peerCh <- routing.PeerInfo{ID: peer.ID(p), Addr: p}
		}
		close(peerCh)
	}()

	return peerCh, nil
}

func (m *MockRouter) Advertise(ctx context.Context, keys []string) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	for _, key := range keys {
		m.resolver[key] = []string{"localhost"}
	}
	return nil
}

func (m *MockRouter) LookupKey(key string) ([]string, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	v, ok := m.resolver[key]
	return v, ok
}
