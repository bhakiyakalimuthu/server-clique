package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/bhakiyakalimuthu/server-clique/types"
	"github.com/go-playground/assert/v2"
	"go.uber.org/zap"
)

func TestServer_Process(t *testing.T) {
	l := zap.NewNop()
	s := NewMemStore(l)

	ctx := context.Background()
	wg := new(sync.WaitGroup)
	w := io.Discard
	cChan := make(chan *types.Message, 1)

	server, err := New(l, w, nil, s, cChan)
	if err != nil {
		t.Fatal("failed to start server", zap.Error(err))
	}

	wg.Add(1)
	go func() {
		server.Process(ctx, wg, 1)
	}()

	t1 := time.Now()
	msg := &types.Message{Action: "add", Key: "111", Value: "222", Timestamp: t1}
	cChan <- msg
	<-time.After(time.Millisecond * 100)
	a1, ok := s.Get(ctx, "111")
	assert.Equal(t, true, ok)
	assert.Equal(t, "222", a1)

	t2 := time.Now()
	_msg := &types.Message{Action: "add", Key: "333", Value: "444", Timestamp: t2}
	cChan <- _msg
	<-time.After(time.Millisecond * 100)
	_actual, ok := s.Get(ctx, "333")
	assert.Equal(t, true, ok)
	assert.Equal(t, "444", _actual)

	_, ok = s.Get(ctx, "ff")
	assert.Equal(t, false, ok)
	<-time.After(time.Millisecond * 100)
	out := s.GetAll(ctx)
	expected := []item{{key: "111", value: "222", timestamp: t1.UnixNano()}, {key: "333", value: "444", timestamp: t2.UnixNano()}}
	assert.Equal(t, expected, out)

	close(cChan) // close the channel to signal the end of messages
	wg.Wait()    // wait for all goroutines to finish
}

func BenchmarkServer_Memstore(b *testing.B) {
	l := zap.NewNop()
	writer := io.Discard
	store := NewMemStore(l)
	cChan := make(chan *types.Message)

	server, err := New(l, writer, nil, store, cChan)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go server.Process(ctx, &wg, i)
	}
	msg := genMessage()
	// Simulate sending messages to the server
	for i := 0; i < b.N; i++ {
		cChan <- msg[i] // create a test message
	}

	close(cChan) // close the channel to signal the end of messages
	wg.Wait()    // wait for all goroutines to finish
}

func BenchmarkServer_MemstoreOptimised(b *testing.B) {
	l := zap.NewNop()
	writer := io.Discard
	store := NewMemStoreOptimised(l)
	cChan := make(chan *types.Message)

	server, err := New(l, writer, nil, store, cChan)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go server.Process(ctx, &wg, i)
	}
	msg := genMessage()
	// Simulate sending messages to the server
	for i := 0; i < b.N; i++ {
		cChan <- msg[i] // create a test message
	}

	close(cChan) // close the channel to signal the end of messages

	wg.Wait() // wait for all goroutines to finish
}

func genMessage() []*types.Message {
	jsonData := `
	[
		{"action": "add","key": "A","value": "a"},
		{"action": "add","key": "B","value": "b"},
		{"action": "add","key": "C","value": "c"},
		{"action": "add","key": "D","value": "d"},
		{"action": "add","key": "E","value": "e"},
		{"action": "unknown_action"},
		{"action": "add","key": "F","value": "f"},
		{"action": "add","key": "G","value": "g"},
		{"action": "add","key": "H","value": "h"},
		{"action": "add","key": "I","value": "i"},
		{"action": "add","key": "J","value": "j"},
		{"action": "add","key": "K","value": "k"},
		{"action": "unknown_action"},
		{"action": "add","key": "L","value": "l"},
		{"action": "add","key": "M","value": "m"},
		{"action": "add","key": "N","value": "n"},
		{"action": "add","key": "O","value": "o"},
		{"action": "getall"},
		{"action": "get","key": "O"},
		{"action": "remove","key": "O"},
		{"action": "get","key": "O"},
		{"action": "getall"},
		{"action": "remove","key": "M"},
		{"action": "remove","key": "N"},
		{"action": "remove","key": "m"},
		{"action": "unknown_action"},
		{"action": "remove","key": "n"},
		{"action": "get","key": "I"},
		{"action": "get","key": "J"},
		{"action": "get","key": "i"},
		{"action": "get","key": "j"},
		{"action": "add","key": "I","value": "ii"},
		{"action": "add","key": "J","value": "jj"},
		{"action": "getall"},
		{"action": "unknown_action"}
	]
	`

	var messages []*types.Message
	err := json.Unmarshal([]byte(jsonData), &messages)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %v", err)
		return nil
	}

	// Generate a high number of test messages
	numMessages := 100000000
	testMessages := make([]*types.Message, numMessages)
	for i := 0; i < numMessages; i++ {
		testMessages[i] = messages[i%len(messages)]
	}
	return testMessages
}
