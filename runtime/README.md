# Runtime packages

`runtime` contains public runtime APIs that generated pzero services may import
directly. Each capability uses a direct package path, such as
`runtime/migrator`. These packages are part of the service compatibility
boundary and should evolve with backward compatibility in mind.

New service-facing runtime capabilities, such as tracing, authentication,
middleware, messaging, configuration, and health checks, should prefer this
layer. Existing `core` packages remain unchanged and can be evaluated
individually instead of being moved as part of an unrelated change.

Generator and CLI implementation code remains under `cmd/pzero`.
