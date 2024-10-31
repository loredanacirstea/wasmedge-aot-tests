package main

import (
	"fmt"
	"os"
	"path"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

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
	ctx := &Context{FilePath: "./wasm/moduleB.wasm", Vm: vmA}
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
	ctx := &Context{Vm: vmA}
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

func runExampleAssemblyScript(dir string) {
	vm := wasmedge.NewVM()
	ctx := &Context{Vm: vm, Calldata: []byte("somecalldata"), Msg: "as wasm example"}
	env := BuildAssemblyScriptEnv(ctx)
	err := vm.RegisterModule(env)
	if err != nil {
		panic(err)
	}
	wasmx := getWasmxEnv(ctx)
	err = vm.RegisterModule(wasmx)
	if err != nil {
		panic(err)
	}
	err = instantiate(vm, path.Join(dir, "as/build/release.wasm"))
	if err != nil {
		panic(err)
	}

	_, err = vm.Execute("main")
	if err != nil {
		panic(err)
	}
	fmt.Println("as/build/release.wasm test success")
}

func runExampleAssemblyScriptAOT(dir string) {
	AotCompile(path.Join(dir, "as/build/release.wasm"), path.Join(dir, "wasm/as.so"))

	vm := wasmedge.NewVM()
	ctx := &Context{Vm: vm, Calldata: []byte("somecalldata"), Msg: "as AOT example"}
	env := BuildAssemblyScriptEnv(ctx)
	err := vm.RegisterModule(env)
	if err != nil {
		panic(err)
	}
	wasmx := getWasmxEnv(ctx)
	err = vm.RegisterModule(wasmx)
	if err != nil {
		panic(err)
	}
	err = instantiate(vm, path.Join(dir, "wasm/as.so"))
	if err != nil {
		panic(err)
	}

	_, err = vm.Execute("main")
	if err != nil {
		panic(err)
	}
	fmt.Println("wasm/as.so test success")
}

func main() {
	wasmedge.SetLogErrorLevel()
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// runExample1(dir)
	// runExampleFibonacci(dir)

	runExampleAssemblyScript(dir)
	runExampleAssemblyScriptAOT(dir)

	fmt.Println("done.")
}
