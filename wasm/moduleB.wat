(module
  (func (export "main") (param $a i32) (result i32)
    (i32.add (local.get $a) (i32.const 13))
  )
)
