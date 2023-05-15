package msg

type SubscribeCh chan []byte

type setSubscribe struct {
	subscriberID string
	subscribeCh  SubscribeCh

	errCh chan error
}

type deleteSubscribe struct {
	subscriberID string

	errCh chan error
}

type (
	setSubscribeCh    chan setSubscribe
	deleteSubscribeCh chan deleteSubscribe
)
