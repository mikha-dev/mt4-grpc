package common

import "mtdealer"

type IDealerLoader interface {
	Load(token string) (*mtdealer.DealerManager, error)
}
