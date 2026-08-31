package util_async

import (
	"errors"

	"github.com/nerago/mopgear-go/util"
)

func GlobalSetErrorFuture(futureError *Future[error], err error) {
	errHandling := futureError.SetResult(err)
	if errHandling != nil {
		util.GlobalFatalErrorHandler(errors.Join(err, errHandling))
	}
}

func GlobalSetErrorFutureResult[T any](futureWithError *FutureCancellableWithError[T], err error) {
	errHandling := futureWithError.SetResultError(err)
	if errHandling != nil {
		util.GlobalFatalErrorHandler(errors.Join(err, errHandling))
	}
}
