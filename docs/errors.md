# Errors

HTTP failures return `jupiter.APIError` with status code, Jupiter code, message, request ID, endpoint, body, and retryability.

Use:

```go
if jupiter.IsRetryable(err) {
	// retry at workflow level if safe
}
```

Transaction and order mutation methods avoid blind automatic retries.

