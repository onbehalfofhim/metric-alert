package retry

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

var ErrBadRequest = errors.New("bad request")

func Retry(fn func() error) error {
	var err error

	for i := 0; i <= len(retryDelays); i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrBadRequest) || !isRetryablePGError(err) {
			return err
		}

		if i < len(retryDelays) {
			time.Sleep(retryDelays[i])
		}
	}

	return fmt.Errorf("all retries failed: %w", err)
}

func isRetryablePGError(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}

	return false
}
