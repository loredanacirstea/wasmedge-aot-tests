package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

var asMemHandler = MemoryHandlerAS{}

// getCallData(): ArrayBuffer
func getCallData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	ptr, err := AllocateWriteMem(ctx.Vm, callframe, ctx.Calldata)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func getWasmxEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i32, getCallData, context, 0))

	return env
}

func wasmxRevert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	return returns, wasmedge.Result_Fail
}

// message: usize, fileName: usize, line: u32, column: u32
func asAbort(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, _ := ReadMemFromPtr(callframe, params[0])
	fileName, _ := ReadMemFromPtr(callframe, params[1])
	fmt.Println(fmt.Sprintf("wasmx_env_1: ABORT: %s, %s. line: %d, column: %d - %s", ReadJsString(message), ReadJsString(fileName), params[2], params[3], ctx.Msg))
	return wasmxRevert(context, callframe, params)
}

func asConsoleLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := asMemHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		fmt.Println(fmt.Sprintf("wasmx: console.log: %s - %s", message, ctx.Msg))
	} else {
		fmt.Println(fmt.Sprintf("wasmx: console.log error: %s - %s", err.Error(), ctx.Msg))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleInfo(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := asMemHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		fmt.Println(fmt.Sprintf("wasmx: console.info: %s - %s", message, ctx.Msg))
	} else {
		fmt.Println(fmt.Sprintf("wasmx: console.info error: %s - %s", err.Error(), ctx.Msg))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleError(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, err := asMemHandler.ReadStringFromPtr(callframe, params[0])
	if err == nil {
		fmt.Println(fmt.Errorf("wasmx: console.error: %s - %s", message, ctx.Msg))
	} else {
		fmt.Println(fmt.Errorf("wasmx: console.error error: %s - %s", err.Error(), ctx.Msg))
	}
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asConsoleDebug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	message, _ := asMemHandler.ReadStringFromPtr(callframe, params[0])
	fmt.Println(fmt.Sprintf("wasmx: console.debug: %s - %s", message, ctx.Msg))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func asDateNow(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = float64(time.Now().UTC().UnixMilli())
	return returns, wasmedge.Result_Success
}

// TODO - move this only for non-deterministic contracts
func asSeed(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = float64(binary.LittleEndian.Uint64(b[:]))
	return returns, wasmedge.Result_Success
}

func BuildAssemblyScriptEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("env")
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__f64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_F64},
	)

	env.AddFunction("abort", wasmedge.NewFunction(functype_i32i32i32i32_, asAbort, context, 0))
	env.AddFunction("console.log", wasmedge.NewFunction(functype_i32_, asConsoleLog, context, 0))
	env.AddFunction("console.info", wasmedge.NewFunction(functype_i32_, asConsoleInfo, context, 0))
	env.AddFunction("console.error", wasmedge.NewFunction(functype_i32_, asConsoleError, context, 0))
	env.AddFunction("console.debug", wasmedge.NewFunction(functype_i32_, asConsoleDebug, context, 0))
	env.AddFunction("Date.now", wasmedge.NewFunction(functype__f64, asDateNow, context, 0))
	env.AddFunction("seed", wasmedge.NewFunction(functype__f64, asSeed, context, 0))
	return env
}

var MEMORY_EXPORT_AS = "__new"

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)

// https://www.assemblyscript.org/runtime.html#memory-layout
// Name	   Offset	Type	Description
// mmInfo	-20	    usize	Memory manager info
// gcInfo	-16	    usize	Garbage collector info
// gcInfo2	-12	    usize	Garbage collector info
// rtId 	-8	    u32	    Unique id of the concrete class
// rtSize	-4	    u32	    Size of the data following the header
//           0		Payload starts here

type MemoryHandlerAS struct{}

func (MemoryHandlerAS) ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	return ReadMemFromPtr(callframe, pointer)
}
func (MemoryHandlerAS) AllocateWriteMem(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	return AllocateWriteMem(vm, callframe, data)
}
func (MemoryHandlerAS) ReadJsString(arr []byte) string {
	return ReadJsString(arr)
}

func (MemoryHandlerAS) ReadStringFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) (string, error) {
	mm, err := ReadMemFromPtr(callframe, pointer)
	if err != nil {
		return "", err
	}
	return ReadJsString(mm), nil
}

func ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	lengthbz, err := ReadMem(callframe, pointer.(int32)-AS_PTR_LENGHT_OFFSET, int32(AS_PTR_LENGHT_OFFSET))
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, err := ReadMem(callframe, pointer, int32(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func AllocateMemVm(vm *wasmedge.VM, size int32) (int32, error) {
	if vm == nil {
		return 0, fmt.Errorf("memory allocation failed, no wasmedge VM instance found")
	}
	result, err := vm.Execute(MEMORY_EXPORT_AS, size, AS_ARRAY_BUFFER_TYPE)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func AllocateWriteMem(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	ptr, err := AllocateMemVm(vm, int32(len(data)))
	if err != nil {
		return ptr, err
	}
	err = WriteMem(callframe, data, ptr)
	if err != nil {
		return ptr, err
	}
	return ptr, nil
}

func ReadJsString(arr []byte) string {
	msg := []byte{}
	for i, char := range arr {
		if i%2 == 0 {
			msg = append(msg, char)
		}
	}
	return string(msg)
}
