package main

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type Context struct {
	FilePath string
	Calldata []byte
	// ContractAddress
}

func instantiate(vm *wasmedge.VM, filePath string) error {
	err := vm.LoadWasmFile(filePath)
	if err != nil {
		return err
	}
	err = vm.Validate()
	if err != nil {
		return err
	}
	err = vm.Instantiate()
	if err != nil {
		return err
	}
	return nil
}

func getCallValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func codeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getCallDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func callDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := readAndFillWithZero(ctx.Calldata, dataStart, dataLen)
	writeMem(callframe, part, params[0])
	return returns, wasmedge.Result_Success
}

func getAddress(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// ctx := context.(*Context)
	// addr := ctx.ContractAddress
	// writeMem(callframe, addr.Bytes(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func getGasLeft(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	returns[0] = int64(100000)
	return returns, wasmedge.Result_Success
}

func revert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Fail
}

func finish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Terminate
}

func writeMem(callframe *wasmedge.CallingFrame, data []byte, pointer interface{}) error {
	ptr := pointer.(int32)
	length := len(data)
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return fmt.Errorf("no memory found")
	}
	err := mem.SetData(data, uint(ptr), uint(length))
	return err
}

func readAndFillWithZero(data []byte, start int32, length int32) []byte {
	dataLen := int32(len(data))
	end := start + length
	var value []byte
	if end >= dataLen {
		if len(data) > 0 {
			value = data[start:]
		}
		value = padWithZeros(value, int(length))
	} else {
		value = data[start:end]
	}
	return value
}

func padWithZeros(data []byte, targetLen int) []byte {
	dataLen := len(data)
	if targetLen <= dataLen {
		return data
	}
	data = append(data, bytes.Repeat([]byte{0}, targetLen-dataLen)...)
	return data
}

func getExternalValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	ctx := context.(*Context)

	vmB := wasmedge.NewVM()
	err := instantiate(vmB, ctx.FilePath)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	res, err := vmB.Execute("main", params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = res[0]
	return returns, wasmedge.Result_Success
}

func getEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("env")
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)

	env.AddFunction("getExternalValue", wasmedge.NewFunction(functype_i32_i32, getExternalValue, context, 0))
	env.AddFunction("ethereum_getCallValue", wasmedge.NewFunction(functype_i32_, getCallValue, context, 0))
	env.AddFunction("ethereum_codeCopy", wasmedge.NewFunction(functype_i32i32i32_, codeCopy, context, 0))
	env.AddFunction("ethereum_getCallDataSize", wasmedge.NewFunction(functype__i32, getCallDataSize, context, 0))
	env.AddFunction("ethereum_callDataCopy", wasmedge.NewFunction(functype_i32i32i32_, callDataCopy, context, 0))
	env.AddFunction("ethereum_getAddress", wasmedge.NewFunction(functype_i32_, getAddress, context, 0))
	env.AddFunction("ethereum_getGasLeft", wasmedge.NewFunction(functype__i64, getGasLeft, context, 0))
	env.AddFunction("ethereum_finish", wasmedge.NewFunction(functype_i32i32_, finish, context, 0))
	env.AddFunction("ethereum_revert", wasmedge.NewFunction(functype_i32i32_, revert, context, 0))

	return env
}

func AotCompile(inPath string, outPath string) error {
	compiler := wasmedge.NewCompiler()
	err := compiler.Compile(inPath, outPath)
	if err != nil {
		fmt.Println("Go: Compile WASM to AOT mode Failed!!")
		return err
	}
	compiler.Release()
	return nil
}

func runExample1(dir string) {
	AotCompile(path.Join(dir, "wasm/moduleA.wasm"), path.Join(dir, "wasm/moduleA.so"))
	AotCompile(path.Join(dir, "wasm/moduleB.wasm"), path.Join(dir, "wasm/moduleB.so"))

	// moduleA
	vmA := wasmedge.NewVM()
	ctx := &Context{FilePath: "./wasm/moduleB.wasm"}
	env := getEnv(ctx)
	err := vmA.RegisterModule(env)
	if err != nil {
		panic(err)
	}
	err = instantiate(vmA, "./wasm/moduleA.so")
	if err != nil {
		panic(err)
	}

	res, err := vmA.Execute("main", int32(4))
	if err != nil {
		panic(err)
	}
	// 4 + 13 + 4
	if res[0].(int32) != 21 {
		panic(fmt.Sprintf("moduleA response should be 21, not %d", res[0]))
	}

	env.Release()
	vmA.Release()
}

func runExampleFibonacci(dir string) {
	AotCompile(path.Join(dir, "wasm/fibonacci.wasm"), path.Join(dir, "wasm/fibonacci.so"))

	// moduleA
	vmA := wasmedge.NewVM()
	ctx := &Context{}
	env := getEnv(ctx)
	err := vmA.RegisterModule(env)
	if err != nil {
		panic(err)
	}
	err = instantiate(vmA, path.Join(dir, "wasm/fibonacci.so"))
	if err != nil {
		panic(err)
	}

	// ctx.ModuleB = vmB

	res, err := vmA.Execute("main", int32(4))
	if err != nil {
		panic(err)
	}
	// 4 + 13 + 4
	fmt.Println("res", res)
}

func main() {
	wasmedge.SetLogErrorLevel()
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	runExample1(dir)
	// runExampleFibonacci(dir)
	fmt.Println("done.")
}
