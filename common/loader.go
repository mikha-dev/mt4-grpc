package common

import (
	"errors"
	"mtdealer"
	"sync"
)

type DealerLoader struct {
	mu      sync.Mutex
	configs map[string]mtdealer.Config
	dealers map[string]*mtdealer.DealerManager
}

func NewDealerLoader(configs map[string]mtdealer.Config) *DealerLoader {
	return &DealerLoader{
		configs: configs,
		dealers: make(map[string]*mtdealer.DealerManager, len(configs)),
	}
}

func (this *DealerLoader) Load(token string) (*mtdealer.DealerManager, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if c, ok := this.configs[token]; !ok {
		return nil, errors.New("ManagerToken not found, check token")
	} else {

		if dealer, ok := this.dealers[token]; ok {
			return dealer, nil
		}

		dealer := mtdealer.NewDealerManager(&c)
		this.dealers[token] = dealer

		return dealer, nil
	}

}

func (this *DealerLoader) Stop() {
	for _, dealer := range this.dealers {
		dealer.Stop()
	}
}
