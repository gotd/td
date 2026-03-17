package downloader

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

const cdnRefreshProbeLimit = 4 * 1024

func (c *cdn) ensureClient(ctx context.Context, dcID int) (CDN, error) {
	c.clientMux.Lock()
	defer c.clientMux.Unlock()

	c.stateMux.RLock()
	current := c.cdn
	currentDC := c.clientDC
	c.stateMux.RUnlock()
	if current != nil && currentDC == dcID {
		return current, nil
	}

	// Redirect may switch DC; recreate client lazily on demand.
	c.closeClient()

	cdnClient, closer, err := c.provider.CDN(ctx, dcID, c.max)
	if err != nil {
		return nil, err
	}
	if cdnClient == nil {
		if closer != nil {
			_ = closer.Close()
		}
		return nil, errors.New("cdn provider returned nil client")
	}

	c.stateMux.Lock()
	c.cdn = cdnClient
	c.closer = closer
	c.clientDC = dcID
	c.stateMux.Unlock()

	return cdnClient, nil
}

func (c *cdn) activateRedirect(ctx context.Context, redirect *tg.UploadFileCDNRedirect) error {
	if redirect == nil {
		c.setMaster()
		return nil
	}
	if _, err := c.ensureClient(ctx, redirect.DCID); err != nil {
		return err
	}
	c.setRedirect(redirect)
	return nil
}

func (c *cdn) recoverCDNControlError(
	ctx context.Context,
	err error,
	offset int64,
	limit int,
	rev uint64,
	loopAttempt int,
) (fallback *chunk, retry bool, handled bool, outErr error) {
	if isCDNFingerprintErr(err) {
		c.closeClient()
		return nil, true, true, nil
	}
	if !isCDNMasterFallbackErr(err) {
		return nil, false, false, nil
	}

	masterChunk, refreshErr := c.refreshRedirect(ctx, offset, limit, rev, loopAttempt)
	if refreshErr != nil {
		return nil, false, true, refreshErr
	}
	if masterChunk != nil {
		return masterChunk, false, true, nil
	}

	return nil, true, true, nil
}

func (c *cdn) refreshRedirect(
	ctx context.Context,
	offset int64,
	limit int,
	prevRev uint64,
	loopAttempt int,
) (*chunk, error) {
	if limit <= 0 {
		limit = cdnRefreshProbeLimit
	}

	c.refreshMux.Lock()
	defer c.refreshMux.Unlock()

	_, _, currentRev := c.snapshot()
	if currentRev != prevRev {
		// Another goroutine already refreshed state; just retry outer loop.
		return nil, nil
	}

	masterChunk, err := retryRequest(
		ctx,
		"refresh CDN redirect",
		func(attempt int, err error) {
			c.reportRetry(RetryOperationRefreshRedirect, attempt, err)
		},
		func() (chunk, error) {
			return c.master.Chunk(ctx, offset, limit)
		},
	)
	if err == nil {
		// Server stopped redirecting this file/token; return to master mode.
		c.setMaster()
		return &masterChunk, nil
	}

	var redirectErr *RedirectError
	if errors.As(err, &redirectErr) {
		if err := c.activateRedirect(ctx, redirectErr.Redirect); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, errors.Wrapf(err, "create CDN client for DC %d", redirectErr.Redirect.DCID)
			}
			if isCDNFingerprintErr(err) {
				c.reportRetry(RetryOperationCreateClient, loopAttempt, err)
				c.closeClient()
				return nil, nil
			}
			return nil, errors.Wrapf(err, "create CDN client for DC %d", redirectErr.Redirect.DCID)
		}
		return nil, nil
	}

	return nil, errors.Wrap(err, "refresh CDN redirect")
}

func (c *cdn) Chunk(ctx context.Context, offset int64, limit int) (chunk, error) {
	// Unified state machine:
	// modeMaster -> try master and switch on redirect;
	// modeCDN    -> serve from CDN with token/keys refresh handling.
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return chunk{}, err
		}

		mode, redirect, rev := c.snapshot()
		switch mode {
		case modeMaster:
			r, err := c.master.Chunk(ctx, offset, limit)
			if err == nil {
				return r, nil
			}

			var redirectErr *RedirectError
			if errors.As(err, &redirectErr) {
				// Redirect is expected protocol path when file is CDN-backed.
				if err := c.activateRedirect(ctx, redirectErr.Redirect); err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return chunk{}, errors.Wrapf(err, "create CDN client for DC %d", redirectErr.Redirect.DCID)
					}
					if isCDNFingerprintErr(err) {
						// CDN keys changed while pool still uses stale keys.
						// Close and retry so provider can reopen with fresh keys.
						c.reportRetry(RetryOperationCreateClient, attempt+1, err)
						c.closeClient()
						continue
					}
					return chunk{}, errors.Wrapf(err, "create CDN client for DC %d", redirectErr.Redirect.DCID)
				}
				continue
			}

			return chunk{}, errors.Wrapf(err, "master chunk offset=%d limit=%d", offset, limit)

		case modeCDN:
			if redirect == nil {
				c.setMaster()
				continue
			}

			cdnClient, err := c.ensureClient(ctx, redirect.DCID)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return chunk{}, errors.Wrapf(err, "create CDN client for DC %d", redirect.DCID)
				}
				if isCDNFingerprintErr(err) {
					// Force recreate with fresh key set.
					c.reportRetry(RetryOperationCreateClient, attempt+1, err)
					c.closeClient()
					continue
				}
				return chunk{}, errors.Wrapf(err, "create CDN client for DC %d", redirect.DCID)
			}

			plan, err := buildCDNRequestPlan(offset, limit)
			if err != nil {
				return chunk{}, errors.Wrapf(err, "cdn request plan offset=%d limit=%d", offset, limit)
			}

			data := make([]byte, 0, limit)
			retryChunk := false

		partLoop:
			for _, req := range plan {
				result, err := cdnClient.UploadGetCDNFile(ctx, &tg.UploadGetCDNFileRequest{
					Offset:    req.offset,
					Limit:     req.limit,
					FileToken: redirect.FileToken,
				})
				if err != nil {
					fallback, retry, handled, recoverErr := c.recoverCDNControlError(ctx, err, offset, limit, rev, attempt+1)
					if recoverErr != nil {
						return chunk{}, recoverErr
					}
					if handled {
						if fallback != nil {
							return *fallback, nil
						}
						if retry {
							c.reportRetry(RetryOperationGetFile, attempt+1, err)
							retryChunk = true
							break partLoop
						}
						continue
					}
					return chunk{}, errors.Wrapf(
						err,
						"cdn chunk dc=%d offset=%d limit=%d",
						redirect.DCID, req.offset, req.limit,
					)
				}

				switch typed := result.(type) {
				case *tg.UploadCDNFile:
					part, err := c.decrypt(typed.Bytes, req.offset, redirect)
					if err != nil {
						return chunk{}, err
					}
					data = append(data, part...)
					if len(part) < req.limit {
						// Reached file tail, remaining plan segments are beyond EOF.
						break partLoop
					}

				case *tg.UploadCDNFileReuploadNeeded:
					// Ask master DC to reissue CDN token window for this file.
					hashes, err := c.client.UploadReuploadCDNFile(ctx, &tg.UploadReuploadCDNFileRequest{
						FileToken:    redirect.FileToken,
						RequestToken: typed.RequestToken,
					})
					if err != nil {
						fallback, retry, handled, recoverErr := c.recoverCDNControlError(ctx, err, offset, limit, rev, attempt+1)
						if recoverErr != nil {
							return chunk{}, recoverErr
						}
						if handled {
							if fallback != nil {
								return *fallback, nil
							}
							if retry {
								c.reportRetry(RetryOperationReupload, attempt+1, err)
								retryChunk = true
								break partLoop
							}
							continue
						}
						return chunk{}, errors.Wrapf(
							err,
							"cdn reupload dc=%d offset=%d limit=%d",
							redirect.DCID, req.offset, req.limit,
						)
					}
					// Reupload returns fresh CDN hashes for the requested token
					// window. Cache them immediately (same strategy as TDesktop) to
					// avoid an extra UploadGetCDNFileHashes call on retry.
					c.cacheHashes(hashes)
					retryChunk = true
					break partLoop

				default:
					return chunk{}, errors.Errorf("unexpected type %T", result)
				}
			}
			if retryChunk {
				continue
			}

			if err := c.verifyChunk(ctx, offset, limit, data); err != nil {
				return chunk{}, err
			}
			return chunk{data: data}, nil
		}
	}

	return chunk{}, retryLimitErr("cdn chunk", maxRetryAttempts, errors.New("state loop"))
}

func (c *cdn) Hashes(ctx context.Context, offset int64) ([]tg.FileHash, error) {
	// Hash retrieval follows same state machine as chunks to stay consistent
	// during concurrent token/redirect changes.
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		mode, redirect, rev := c.snapshot()
		switch mode {
		case modeMaster:
			hashes, err := c.master.Hashes(ctx, offset)
			if err != nil {
				return nil, errors.Wrapf(err, "master hashes offset=%d", offset)
			}
			return hashes, nil

		case modeCDN:
			if redirect == nil {
				c.setMaster()
				continue
			}

			hashes, err := c.client.UploadGetCDNFileHashes(ctx, &tg.UploadGetCDNFileHashesRequest{
				FileToken: redirect.FileToken,
				Offset:    offset,
			})
			if err != nil {
				_, retry, handled, recoverErr := c.recoverCDNControlError(ctx, err, offset, cdnRefreshProbeLimit, rev, attempt+1)
				if recoverErr != nil {
					return nil, recoverErr
				}
				if handled && retry {
					c.reportRetry(RetryOperationGetFileHashes, attempt+1, err)
					continue
				}
				if handled {
					continue
				}
				return nil, errors.Wrapf(err, "cdn hashes dc=%d offset=%d", redirect.DCID, offset)
			}
			c.cacheHashes(hashes)
			return hashes, nil
		}
	}

	return nil, retryLimitErr("cdn hashes", maxRetryAttempts, errors.New("state loop"))
}
