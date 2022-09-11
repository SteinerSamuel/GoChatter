type User struct {
	name           string
	channelHandler nil

	stopListenerChan chan struct{}
	listening        bool

	messageChan chan struct{}
}