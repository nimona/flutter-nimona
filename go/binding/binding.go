package main

//#include <stdint.h>
//#include <stdlib.h>
//typedef struct { void* message; int size; char* error; } BytesReturn;
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/rs/xid"

	"flutter.nimona.io/provider"

	"nimona.io/pkg/context"
	"nimona.io/pkg/object"
)

var (
	nimonaProvider     *provider.Provider
	subscriptionsMutex sync.RWMutex
	subscriptions      map[string]object.ReadCloser
)

func init() {
	if nimonaProvider != nil {
		return
	}
	nimonaProvider = provider.New()
	subscriptionsMutex = sync.RWMutex{}
	subscriptions = map[string]object.ReadCloser{}
}

func renderBytes(b []byte, err error) *C.BytesReturn {
	r := (*C.BytesReturn)(C.malloc(C.size_t(C.sizeof_BytesReturn)))
	if err != nil {
		r.error = C.CString(err.Error())
		return r
	}
	r.error = nil
	r.message = C.CBytes(b)
	r.size = C.int(len(b))
	return r
}

func renderObject(o *object.Object) *C.BytesReturn {
	m := o.ToMap()
	m["_hash:s"] = o.Hash().String()
	b, err := json.Marshal(m)
	return renderBytes(b, err)
}

//export NimonaBridgeCall
func NimonaBridgeCall(
	name *C.char,
	payload unsafe.Pointer,
	payloadSize C.int,
) *C.BytesReturn {
	nameString := C.GoString(name)
	payloadBytes := C.GoBytes(payload, payloadSize)

	fmt.Printf("> Called %s with %s\n", nameString, string(payloadBytes))

	switch nameString {
	case "get":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		req := provider.GetRequest{}
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			return renderBytes(nil, err)
		}
		r, err := nimonaProvider.Get(ctx, req)
		if err != nil {
			return renderBytes(nil, err)
		}
		key := xid.New().String()
		subscriptionsMutex.Lock()
		subscriptions[key] = r
		subscriptionsMutex.Unlock()
		return renderBytes([]byte(key), nil)
	case "subscribe":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		r, err := nimonaProvider.Subscribe(ctx, string(payloadBytes))
		if err != nil {
			return renderBytes(nil, err)
		}
		key := xid.New().String()
		subscriptionsMutex.Lock()
		subscriptions[key] = r
		subscriptionsMutex.Unlock()
		return renderBytes([]byte(key), nil)
	case "pop":
		subscriptionsMutex.RLock()
		r, ok := subscriptions[string(payloadBytes)]
		if !ok {
			return renderBytes(nil, errors.New("missing subscription key"))
		}
		subscriptionsMutex.RUnlock()
		o, err := r.Read()
		if err != nil {
			return renderBytes(nil, err)
		}
		return renderObject(o)
	case "cancel":
		subscriptionsMutex.Lock()
		r, ok := subscriptions[string(payloadBytes)]
		if !ok {
			return renderBytes(nil, errors.New("missing subscription key"))
		}
		subscriptionsMutex.Unlock()
		r.Close()
		delete(subscriptions, string(payloadBytes))
		return renderBytes(nil, nil)
	case "requestStream":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		r, err := nimonaProvider.RequestStream(
			ctx,
			object.Hash(string(payloadBytes)),
		)
		if err != nil {
			return renderBytes(nil, err)
		}
		go object.ReadAll(r)
		return renderBytes(nil, nil)
	case "put":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		m := map[string]interface{}{}
		if err := json.Unmarshal(payloadBytes, &m); err != nil {
			return renderBytes(nil, err)
		}
		o := object.FromMap(m)
		u, err := nimonaProvider.Put(ctx, o)
		if err != nil {
			return renderBytes(nil, err)
		}
		return renderObject(u)
	case "getFeedRootHash":
		feedRootHash := nimonaProvider.GetFeedRootHash(string(payloadBytes))
		return renderBytes([]byte(feedRootHash), nil)
	}

	return renderBytes([]byte("error"), errors.New(nameString+" not implemented"))
}

// Unused
func main() {}
