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
	"nimona.io/pkg/version"
)

var (
	nimonaProvider     *provider.Provider
	subscriptionsMutex sync.RWMutex
	subscriptions      map[string]object.ReadCloser
)

func renderBytes(b []byte, err error) *C.BytesReturn {
	r := (*C.BytesReturn)(C.malloc(C.size_t(C.sizeof_BytesReturn)))
	if err != nil {
		fmt.Println("++ ERROR", err)
		r.error = C.CString(err.Error())
		return r
	}
	r.error = nil
	r.message = C.CBytes(b)
	r.size = C.int(len(b))
	return r
}

func marshalObject(o *object.Object) ([]byte, error) {
	m := object.Copy(o).ToMap()
	m["_hash:s"] = o.Hash().String()
	return json.Marshal(m)
}

func renderObject(o *object.Object) *C.BytesReturn {
	b, err := marshalObject(o)
	fmt.Println("++ RESP body=", string(string(b)))
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
	case "init":
		if nimonaProvider != nil {
			return renderBytes(nil, nil)
		}
		nimonaProvider = provider.New()
		subscriptionsMutex = sync.RWMutex{}
		subscriptions = map[string]object.ReadCloser{}
		return renderBytes(nil, nil)
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
			fmt.Println("++ ERROR", err)
			return renderBytes(nil, err)
		}
		os := []string{}
		for {
			o, err := r.Read()
			if err != nil || o == nil {
				break
			}
			b, err := marshalObject(o)
			if err != nil {
				fmt.Println("++ ERROR", err)
				return renderBytes(nil, err)
			}
			os = append(os, string(b))
		}
		res := &provider.GetResponse{
			ObjectBodies: os,
		}
		b, err := json.Marshal(res)
		fmt.Println("++ RESP body=", string(b))
		return renderBytes(b, err)
	case "version":
		fmt.Println("++ RESP version=", version.Version)
		return renderBytes([]byte(version.Version), nil)
	case "subscribe":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		r, err := nimonaProvider.Subscribe(ctx, string(payloadBytes))
		if err != nil {
			fmt.Println("++ ERROR", err)
			return renderBytes(nil, err)
		}
		key := xid.New().String()
		subscriptionsMutex.Lock()
		subscriptions[key] = r
		subscriptionsMutex.Unlock()
		fmt.Println("++ RESP key=", key)
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
			fmt.Println("++ ERROR", err)
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
			context.WithTimeout(10 * time.Second),
		)
		if err := nimonaProvider.RequestStream(
			ctx,
			object.Hash(string(payloadBytes)),
		); err != nil {
			return renderBytes(nil, err)
		}
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
			fmt.Println("++ ERROR", err)
			return renderBytes(nil, err)
		}
		return renderObject(u)
	case "getFeedRootHash":
		feedRootHash := nimonaProvider.GetFeedRootHash(string(payloadBytes))
		return renderBytes([]byte(feedRootHash), nil)
	case "getConnectionInfo":
		o := nimonaProvider.GetConnectionInfo().ToObject()
		return renderObject(o)
	}

	return renderBytes([]byte("error"), errors.New(nameString+" not implemented"))
}

// Unused
func main() {}
