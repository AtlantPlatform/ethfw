package ethfw

import "time"

type clientOpt func(o *clientOptions)

type clientOptions struct {
	accountSyncTimeout time.Duration
	chainID            int64
}

func TimeoutOpt(timeout time.Duration) clientOpt {
	return func(o *clientOptions) {
		o.accountSyncTimeout = timeout
	}
}

func ChainOpt(id int64) clientOpt {
	return func(o *clientOptions) {
		o.chainID = id
	}
}
