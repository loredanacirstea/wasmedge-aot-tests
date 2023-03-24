package main

import (
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type Context struct {
	ModuleB *wasmedge.VM
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

func getExternalValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	ctx := context.(*Context)
	res, err := ctx.ModuleB.Execute("main", params[0])
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
	env.AddFunction("getExternalValue", wasmedge.NewFunction(functype_i32_i32, getExternalValue, context, 0))
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

func main() {
	wasmedge.SetLogErrorLevel()

	AotCompile("./wasm/moduleA.wasm", "./wasm/moduleA.so")
	AotCompile("./wasm/moduleB.wasm", "./wasm/moduleB.so")

	// moduleA
	vmA := wasmedge.NewVM()
	ctx := &Context{}
	env := getEnv(ctx)
	err := vmA.RegisterModule(env)
	if err != nil {
		panic(err)
	}
	err = instantiate(vmA, "./wasm/moduleA.so")
	if err != nil {
		panic(err)
	}

	// moduleB
	vmB := wasmedge.NewVM()
	err = instantiate(vmB, "./wasm/moduleB.wasm")
	if err != nil {
		panic(err)
	}

	ctx.ModuleB = vmB

	res, err := vmA.Execute("main", int32(4))
	if err != nil {
		panic(err)
	}
	// 4 + 13 + 4
	fmt.Println("res", res)
}
