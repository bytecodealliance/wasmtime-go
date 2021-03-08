(module
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory 1)
  (export "memory" (memory 0))
  (data (i32.const 8) "hello world\n")
  (func $main (export "_start")
    (i32.store (i32.const 0) (i32.const 8))  ;; iov.iov_base
    (i32.store (i32.const 4) (i32.const 12))  ;; iov.iov_len
    (call $fd_write
      (i32.const 1)  ;; file_descriptor
      (i32.const 0)  ;; *iovs
      (i32.const 1)  ;; iovs_len
      (i32.const 20)  ;; nwritten
    )
    drop
  )
)
