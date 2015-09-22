package pubsub

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

// A pub-sub message - defined to support Redis receiving different
// message types, such as subscribe/unsubscribe info.
type Message struct {
	Type    string
	Channel string
	Data    string
}

// Client interface for both Redis and ZMQ pubsub clients.
type Client interface {
	Subscribe(channels ...interface{}) (err error)
	Unsubscribe(channels ...interface{}) (err error)
	Publish(channel string, message string)
	Receive() (message Message)
}

// Redis client - defines the underlying connection and pub-sub
// connections, as well as a mutex for locking write access,
// since this occurs from multiple goroutines.
type RedisClient struct {
	conn redis.Conn
	redis.PubSubConn
	sync.Mutex
}

// Returns a new Redis client. The underlying redigo package uses
// Go's bufio package which will flush the connection when it contains
// enough data to send, but we still need to set up some kind of timed
// flusher, so it's done here with a goroutine.
func NewRedisClient(host string) *RedisClient {
	host = fmt.Sprintf("%s:6379", host)
	conn, _ := redis.Dial("tcp", host)
	pubsub, _ := redis.Dial("tcp", host)
	client := RedisClient{conn, redis.PubSubConn{pubsub}, sync.Mutex{}}
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			client.Lock()
			client.conn.Flush()
			client.Unlock()
		}
	}()
	return &client
}

func (client *RedisClient) Publish(channel, message string) {
	client.Lock()
	client.conn.Send("PUBLISH", channel, message)
	client.Unlock()
}

func (client *RedisClient) Receive() Message {
	switch message := client.PubSubConn.Receive().(type) {
	case redis.Message:
		return Message{"message", message.Channel, string(message.Data)}
	case redis.Subscription:
		return Message{message.Kind, message.Channel, string(message.Count)}
	}
	return Message{}
}
