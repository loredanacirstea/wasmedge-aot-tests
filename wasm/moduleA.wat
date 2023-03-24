(module
  (import "env" "getExternalValue" (func $getExternalValue (param i32) (result i32)))
  (func (export "main") (param $a i32) (result i32)
    (i32.add (local.get $a) (call $getExternalValue (local.get $a)))
  )
)
