import * as wasmx from "./wasmx";

export function wasmx_env_2(): void {}

export function instantiate(): void {}

export function main(): void {
  console.log("--main--")
  const calldraw = wasmx.getCallData();
  console.log("--calldraw--")
  let calldstr = String.UTF8.decode(calldraw)
  console.log("--calldstr--" + calldstr)
}
