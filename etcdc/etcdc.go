package etcdc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const optimeout time.Duration = 5

type etcdConn struct {
	addresses []string
	username  string
	password  string
	client    *clientv3.Client
}

// NewC init a etcd client
func NewC(addresses []string, username, password string) (*etcdConn, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   addresses,
		Username:    username,
		Password:    password,
		DialTimeout: optimeout * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &etcdConn{addresses, username, password, c}, nil
}

func (c *etcdConn) Close() error {
	return c.client.Close()
}

// CreateNX create a key with value if the key does not exist
func (c *etcdConn) CreateNX(key, value string) error {
	kvc := clientv3.NewKV(c.client)

	ctx, cancel := context.WithTimeout(context.Background(), optimeout*time.Second)
	defer cancel()

	_, err := kvc.Txn(ctx).
		If(
			clientv3.Compare(clientv3.Version(key), "=", 0),
		).
		Then(clientv3.OpPut(key, value)).
		Commit()
	return err
}

// Update update key value
func (c *etcdConn) Update(key, value string, lease clientv3.LeaseID) error {
	kvc := clientv3.NewKV(c.client)

	ctx, cancel := context.WithTimeout(context.Background(), optimeout*time.Second)
	defer cancel()

	var err error
	if lease == clientv3.LeaseID(0) {
		_, err = kvc.Put(ctx, key, value)
	} else {
		_, err = kvc.Put(ctx, key, value, clientv3.WithLease(lease))
	}
	return err
}

// Get query and get
func (c *etcdConn) Get(pattern string, isPrefix bool, keyOnly bool) ([]*mvccpb.KeyValue, error) {
	kvc := clientv3.NewKV(c.client)

	ops := []clientv3.OpOption{}
	if isPrefix {
		ops = append(ops, clientv3.WithPrefix())
	}
	if keyOnly {
		ops = append(ops, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(context.Background(), optimeout*time.Second)
	defer cancel()
	resp, err := kvc.Get(ctx, pattern, ops...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, nil
}

// KeepAlive keep alive
func (c *etcdConn) KeepAlive(lease clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), optimeout*time.Second)
	// defer cancel()

	// timeout or deadline context should not be used for lease keep alive
	ch, err := c.client.KeepAlive(context.Background(), lease)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// CreateLease create a new lease
func (c *etcdConn) CreateLease(ttl int64) (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), optimeout*time.Second)
	defer cancel()

	resp, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// Encode encode an object to a string
func Encode(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	s := base64.StdEncoding.EncodeToString(b)
	return s, nil
}

// Decode decode a string into an object
func Decode(s string, v interface{}) error {

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}
