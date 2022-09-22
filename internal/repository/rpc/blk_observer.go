/*
Package rpc implements bridge to Lachesis full node API interface.

We recommend using local IPC for fast and the most efficient inter-process communication between the API server
and an Opera/Lachesis node. Any remote RPC connection will work, but the performance may be significantly degraded
by extra networking overhead of remote RPC calls.

You should also consider security implications of opening Lachesis RPC interface for remote access.
If you considering it as your deployment strategy, you should establish encrypted channel between the API server
and Lachesis RPC interface with connection limited to specified endpoints.

We strongly discourage opening Lachesis RPC interface for unrestricted Internet access.
*/
package rpc

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
)

// glxyHeadsObserverSubscribeTick represents the time between subscription attempts.
const glxyHeadsObserverSubscribeTick = 30 * time.Second

// observeBlocks collects new blocks from the blockchain network
// and posts them into the proxy channel for processing.
func (glxy *GlxyBridge) observeBlocks() {
	var sub ethereum.Subscription
	defer func() {
		if sub != nil {
			sub.Unsubscribe()
		}
		glxy.log.Noticef("block observer done")
		glxy.wg.Done()
	}()

	sub = glxy.blockSubscription()
	for {
		// re-subscribe if the subscription ref is not valid
		if sub == nil {
			tm := time.NewTimer(glxyHeadsObserverSubscribeTick)
			select {
			case <-glxy.sigClose:
				return
			case <-tm.C:
				sub = glxy.blockSubscription()
				continue
			}
		}

		// use the subscriptions
		select {
		case <-glxy.sigClose:
			return
		case err := <-sub.Err():
			glxy.log.Errorf("block subscription failed; %s", err.Error())
			sub = nil
		}
	}
}

// blockSubscription provides a subscription for new blocks received
// by the connected blockchain node.
func (glxy *GlxyBridge) blockSubscription() ethereum.Subscription {
	sub, err := glxy.rpc.EthSubscribe(context.Background(), glxy.headers, "newHeads")
	if err != nil {
		glxy.log.Criticalf("can not observe new blocks; %s", err.Error())
		return nil
	}
	return sub
}
