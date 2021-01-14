package main

//#include <stdint.h>
//#include <stdlib.h>
//typedef struct { void* message; int size; char* error; } BytesReturn;
import "C"

import (
	"errors"
	"fmt"
	"time"
	"unsafe"
)

var ch = make(chan string, 100)

//export NimonaBridgeCall
func NimonaBridgeCall(
	name *C.char,
	payload unsafe.Pointer,
	payloadSize C.int,
) *C.BytesReturn {
	marshal := func(b []byte, err error) *C.BytesReturn {
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

	nameString := C.GoString(name)
	payloadBytes := C.GoBytes(payload, payloadSize)

	fmt.Printf("> Called %s with %s\n", nameString, string(payloadBytes))

	switch nameString {
	case "pop":
		s := <-ch
		return marshal([]byte(s), nil)
	case "subscribe":
		go func() {
			t := time.NewTicker(5 * time.Second)
			i := 0
			for {
				select {
				case <-t.C:
					ch <- fmt.Sprintf("hello world %d", i)
				}
				i++
			}
		}()
		return marshal([]byte("ok"), nil)
	}

	return marshal([]byte("error"), errors.New(nameString+" not implemented"))
}

// Unused
func main() {}
