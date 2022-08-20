---functions---

// Custom gotd tracing wrapper. See https://w3c.github.io/trace-context/, https://www.w3.org/TR/baggage/
invokeWithTrace#104be9ab {X:Type} flags:# traceParent:flags.1?string traceState:flags.2?string baggage:flags.3?string query:!X = X;
