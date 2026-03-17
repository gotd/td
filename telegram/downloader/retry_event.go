package downloader

// RetryEvent describes retried transient downloader error.
type RetryEvent struct {
	// Operation identifies retry source.
	Operation string
	// Attempt is 1-based counter inside current retry loop.
	Attempt int
	// Err is the error that triggered retry.
	Err error
}

const (
	RetryOperationGetFile         = "cdn.get_file"
	RetryOperationGetFileHashes   = "cdn.get_file_hashes"
	RetryOperationReupload        = "cdn.reupload"
	RetryOperationRefreshRedirect = "cdn.refresh_redirect"
	RetryOperationCreateClient    = "cdn.create_client"
	RetryOperationReaderChunk     = "reader.chunk"
	RetryOperationReaderHashes    = "reader.hashes"
)

// RetryHandler is called for every downloader error that is retried internally.
type RetryHandler func(event RetryEvent)

type retryReporter interface {
	reportRetry(operation string, attempt int, err error)
}

func reportSchemaRetry(s schema, operation string, attempt int, err error) {
	if attempt < 1 || err == nil {
		return
	}

	reporter, ok := s.(retryReporter)
	if !ok {
		return
	}

	reporter.reportRetry(operation, attempt, err)
}
