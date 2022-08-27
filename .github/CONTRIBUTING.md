# Contributing

## TODOs

- Syntactic sugar for refs
- Review patch semantics
- Add error types for returning status codes
- Generate a TS client
- Introspection API
- Support default values from the schema
- Real CLI w/ distribution
- Generate a Go client

### Schema

- Parse comments from JSON5 as descriptions

### Go server

- Improve generated names
- Handle path-clashes / verbs
- Handle non-string GET parameters
- Translate JTD errors into English
- Timeouts, Accept-Encoding, ...?
- Validate Content-Type
- Handle panics
- Benchmark

### Ideas

- Improved invalid schema validation errors
- Inline validation for editing JSON5+JTD files
- Response header with warnings, e.g. unknown field names
- Encode warnings in error responses
- Support url-form-encoded: https://brandur.org/fragments/application-x-wwww-form-urlencoded
- Generate a TS server
- Support streaming via SSE
- OpenAPI integration, e.g. Postman
- Built-in pagination support
- https://brandur.org/idempotency-keys / https://brandur.org/fragments/is-transient / https://brandur.org/fragments/idempotency-keys-crunchy
- Linting (e.g. "REST"-like, consistent casing, ...)
- Basic doc generator
- Review https://brandur.org/fragments/openai-api - OpenAI, Stripe, Crunchy
  - https://brandur.org/fragments
- Encode real examples/