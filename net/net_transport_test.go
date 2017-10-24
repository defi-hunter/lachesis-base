package net

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/babbleio/babble/common"
	"github.com/babbleio/babble/hashgraph"
)

func TestNetworkTransport_StartStop(t *testing.T) {
	trans, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	trans.Close()
}

func TestNetworkTransport_Sync(t *testing.T) {
	// Transport 1 is consumer
	trans1, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := SyncRequest{
		From: "A",
		Known: map[int]int{
			0: 1,
			1: 2,
			2: 3,
		},
	}
	resp := SyncResponse{
		From: "B",
		Events: []hashgraph.WireEvent{
			hashgraph.WireEvent{
				Body: hashgraph.WireBody{
					Transactions:         [][]byte(nil),
					SelfParentIndex:      1,
					OtherParentCreatorID: 10,
					OtherParentIndex:     0,
					CreatorID:            9,
				},
			},
		},
		Known: map[int]int{
			0: 5,
			1: 5,
			2: 6,
		},
	}

	// Listen for a request
	go func() {
		select {
		case rpc := <-rpcCh:
			// Verify the command
			req := rpc.Command.(*SyncRequest)
			if !reflect.DeepEqual(req, &args) {
				t.Fatalf("command mismatch: %#v %#v", *req, args)
			}

			rpc.Respond(&resp, nil)

		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	// Transport 2 makes outbound request
	trans2, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	var out SyncResponse
	if err := trans2.Sync(trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !reflect.DeepEqual(resp, out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}
}

func TestNetworkTransport_EagerSync(t *testing.T) {
	// Transport 1 is consumer
	trans1, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := EagerSyncRequest{
		From: "A",
		Events: []hashgraph.WireEvent{
			hashgraph.WireEvent{
				Body: hashgraph.WireBody{
					Transactions:         [][]byte(nil),
					SelfParentIndex:      1,
					OtherParentCreatorID: 10,
					OtherParentIndex:     0,
					CreatorID:            9,
				},
			},
		},
	}
	resp := EagerSyncResponse{
		Success: true,
	}

	// Listen for a request
	go func() {
		select {
		case rpc := <-rpcCh:
			// Verify the command
			req := rpc.Command.(*EagerSyncRequest)
			if !reflect.DeepEqual(req, &args) {
				t.Fatalf("command mismatch: %#v %#v", *req, args)
			}

			rpc.Respond(&resp, nil)

		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	// Transport 2 makes outbound request
	trans2, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	var out EagerSyncResponse
	if err := trans2.EagerSync(trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !reflect.DeepEqual(resp, out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}
}

func TestNetworkTransport_FastForward(t *testing.T) {
	// Transport 1 is consumer
	trans1, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := FastForwardRequest{
		From: "A",
	}
	resp := FastForwardResponse{
		From: "B",
		Frame: hashgraph.Frame{
			Roots: map[string]hashgraph.Root{
				"0": hashgraph.Root{
					X:     "x0",
					Y:     "y0",
					Index: 4,
					Round: 2,
					Others: map[string]string{
						"o1": "oldEvent",
					},
				},
				"1": hashgraph.Root{
					X:     "x1",
					Y:     "y1",
					Index: 4,
					Round: 2,
				},
				"2": hashgraph.Root{
					X:     "x2",
					Y:     "y2",
					Index: 4,
					Round: 2,
				},
			},
			Events: []hashgraph.Event{
				hashgraph.Event{
					Body: hashgraph.EventBody{
						Transactions: [][]byte(nil),
						Parents:      []string{"p1", "p2"},
						Creator:      []byte("creator"),
						Index:        19,
						Timestamp:    time.Now().UTC(),
					},
				},
			},
		},
	}

	// Listen for a request
	go func() {
		select {
		case rpc := <-rpcCh:
			// Verify the command
			req := rpc.Command.(*FastForwardRequest)
			if !reflect.DeepEqual(req, &args) {
				t.Fatalf("command mismatch: %#v %#v", *req, args)
			}

			rpc.Respond(&resp, nil)

		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	// Transport 2 makes outbound request
	trans2, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	var out FastForwardResponse
	if err := trans2.FastForward(trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !reflect.DeepEqual(resp, out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}
}

func TestNetworkTransport_PooledConn(t *testing.T) {
	// Transport 1 is consumer
	trans1, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := SyncRequest{
		From: "A",
		Known: map[int]int{
			0: 1,
			1: 2,
			2: 3,
		},
	}
	resp := SyncResponse{
		From: "B",
		Events: []hashgraph.WireEvent{
			hashgraph.WireEvent{
				Body: hashgraph.WireBody{
					Transactions:         [][]byte(nil),
					SelfParentIndex:      1,
					OtherParentCreatorID: 10,
					OtherParentIndex:     0,
					CreatorID:            9,
				},
			},
		},
	}

	// Listen for a request
	go func() {
		for {
			select {
			case rpc := <-rpcCh:
				// Verify the command
				req := rpc.Command.(*SyncRequest)
				if !reflect.DeepEqual(req, &args) {
					t.Fatalf("command mismatch: %#v %#v", *req, args)
				}
				rpc.Respond(&resp, nil)

			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()

	// Transport 2 makes outbound request, 3 conn pool
	trans2, err := NewTCPTransport("127.0.0.1:0", nil, 3, time.Second, common.NewTestLogger(t))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	// Create wait group
	wg := &sync.WaitGroup{}
	wg.Add(5)

	appendFunc := func() {
		defer wg.Done()
		var out SyncResponse
		if err := trans2.Sync(trans1.LocalAddr(), &args, &out); err != nil {
			t.Fatalf("err: %v", err)
		}

		// Verify the response
		if !reflect.DeepEqual(resp, out) {
			t.Fatalf("command mismatch: %#v %#v", resp, out)
		}
	}

	// Try to do parallel appends, should stress the conn pool
	for i := 0; i < 5; i++ {
		go appendFunc()
	}

	// Wait for the routines to finish
	wg.Wait()

	// Check the conn pool size
	addr := trans1.LocalAddr()
	if len(trans2.connPool[addr]) != 3 {
		t.Fatalf("Expected 2 pooled conns!")
	}
}
