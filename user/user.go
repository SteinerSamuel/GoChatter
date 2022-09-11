package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
)

const (
	userKey        = "users"
	userChannelFmt = "user%s:channels"
	ChannelsKey    = "channels"
)

var ctx = context.Background()

type User struct {
	name           string
	channelHandler *redis.PubSub

	stopListenerChan chan struct{}
	listening        bool

	messageChan chan redis.Message
}

func Connect(redisdb *redis.Client, name string) (*User, error) {
	if _, err := redisdb.SAdd(ctx, userKey, name).Result(); err != nil {
		return nil, err
	}

	u := &User{
		name:             name,
		stopListenerChan: make(chan struct{}),
		messageChan:      make(chan redis.Message),
	}

	if err := u.connect(redisdb); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Subscribe(redisdb *redis.Client, channel string) error {
	userChannelsKey := fmt.Sprintf(userChannelFmt, u.name)

	if redisdb.SIsMember(ctx, userChannelsKey, channel).Val() {
		return nil
	}
	if err := redisdb.SAdd(ctx, userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect(redisdb)
}

func (u *User) Unsubscribe(redisdb *redis.Client, channel string) error {
	userChannelsKey := fmt.Sprintf(userChannelFmt, u.name)

	if !redisdb.SIsMember(ctx, userChannelsKey, channel).Val() {
		return nil
	}
	if err := redisdb.SRem(ctx, userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect(redisdb)
}

func (u *User) connect(redisdb *redis.Client) error {
	var c []string

	c1, err := redisdb.SMembers(ctx, ChannelsKey).Result()
	if err != nil {
		return err
	}
	c = append(c, c1...)

	// get all user channels (from DB) and start subscribe
	c2, err := redisdb.SMembers(ctx, fmt.Sprintf(userChannelFmt, u.name)).Result()
	if err != nil {
		return err
	}
	c = append(c, c2...)

	if len(c) == 0 {
		fmt.Println("no channels to connect to for user: ", u.name)
		return nil
	}

	if u.channelHandler != nil {
		if err := u.channelHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := u.channelHandler.Close(); err != nil {
			return err
		}
	}

	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	return u.doConnect(redisdb, c...)
}

func (u *User) doConnect(redisdb *redis.Client, channels ...string) error {
	// subscribe all channels in one request
	pubSub := redisdb.Subscribe(ctx, channels...)
	// keep channel handler to be used in unsubscribe
	u.channelHandler = pubSub

	// The Listener
	go func() {
		u.listening = true
		fmt.Println("starting the listener for user:", u.name, "on channels:", channels)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				u.messageChan <- *msg

			case <-u.stopListenerChan:
				fmt.Println("stopping the listener for user:", u.name)
				return
			}
		}
	}()
	return nil
}

func (u *User) Disconnect() error {
	if u.channelHandler != nil {
		if err := u.channelHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := u.channelHandler.Close(); err != nil {
			return err
		}
	}
	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	close(u.messageChan)

	return nil
}

func Chat(rdb *redis.Client, channel string, content string) error {
	return rdb.Publish(ctx, channel, content).Err()
}

func List(rdb *redis.Client) ([]string, error) {
	return rdb.SMembers(ctx, userKey).Result()
}

func GetChannels(rdb *redis.Client, username string) ([]string, error) {

	if !rdb.SIsMember(ctx, userKey, username).Val() {
		return nil, errors.New("user not exists")
	}

	var c []string

	c1, err := rdb.SMembers(ctx, ChannelsKey).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c1...)

	// get all user channels (from DB) and start subscribe
	c2, err := rdb.SMembers(ctx, fmt.Sprintf(userChannelFmt, username)).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c2...)

	return c, nil
}
