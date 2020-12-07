package etcdc

import (
	"os"
	"testing"
)

var (
	address  string = "localhost:2379"
	user     string = "root"
	password string = "password"
)

func init() {
	ep, ok := os.LookupEnv("ETCD_ENDPOINT")
	if ok {
		address = ep
	}

	usr, ok := os.LookupEnv("ETCD_USER")
	if ok {
		user = usr
	}

	passwd, ok := os.LookupEnv("ETCD_PASSWORD")
	if ok {
		password = passwd
	}
}

func makeConn() *etcdConn {
	c, err := NewC([]string{address}, user, password)
	if err != nil {
		panic(err)
	}
	return c
}

func TestNewC(t *testing.T) {
	c := makeConn()
	defer c.Close()
}

func TestCreateNX(t *testing.T) {
	data := []struct {
		k string
		v string
	}{
		{"foo", "100"},
		{"foo", "200"},
		{"bar", "100"},
		{"bar", "200"},
	}

	c := makeConn()
	defer c.Close()

	for _, pair := range data {
		err := c.CreateNX(pair.k, pair.v)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestUpdate(t *testing.T) {
	data := []struct {
		k string
		v string
	}{
		{"foo", "1000"},
		{"bar", "1000"},
	}

	c := makeConn()
	defer c.Close()

	for _, pair := range data {
		err := c.Update(pair.k, pair.v, 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	id, err := c.CreateLease(30)
	if err != nil {
		t.Fatal(err)
	}

	for _, pair := range data {
		err := c.Update(pair.k, pair.v, id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGet(t *testing.T) {
	data := []struct {
		k string
		v string
	}{
		{"foo", "100"},
		{"foo1", "300"},
		{"foo2", "500"},
	}

	c := makeConn()
	defer c.Close()

	for _, pair := range data {
		err := c.Update(pair.k, pair.v, 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, isPrefix := range []bool{false, true} {
		for _, keyOnly := range []bool{false, true} {
			kvs, err := c.Get("foo", isPrefix, keyOnly)
			if err != nil {
				t.Logf("Fail to get key foo: %v", err)
				t.Fail()
			}
			for _, kv := range kvs {
				t.Logf("KV: %v", kv)
			}
		}
	}
}

func TestKeepAlive(t *testing.T) {
	c := makeConn()
	defer c.Close()

	lease, err := c.CreateLease(30)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Create lease %v", lease)

	err = c.Update("foo", "100", lease)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Update key foo with lease %v", lease)

	lch, err := c.KeepAlive(lease)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		resp := <-lch
		t.Logf("Keep alive lease %v with TTL %v", resp.ID, resp.TTL)
	}
	t.Logf("Aftet function return, the client will be closed and hence the lease %v should expire soon", lease)
}

func TestCreateLease(t *testing.T) {
	ttl := 30
	c := makeConn()
	defer c.Close()

	id, err := c.CreateLease(int64(ttl))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Create lease: %v", id)
}

func TestEncodeDecode(t *testing.T) {
	feeds := []map[string]interface{}{
		{"k11": "v11"},
		{"k21": 1},
		{"k31": map[string]interface{}{"k311": "v311", "k312": 1}},
	}

	eresults := []string{}
	for _, v := range feeds {
		s, err := Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		eresults = append(eresults, s)
		t.Logf("Original data: %v\n", v)
		t.Logf("Encode result: %s\n", s)
	}

	for _, v := range eresults {
		target := map[string]interface{}{}
		err := Decode(v, &target)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("original data: %v\n", v)
		t.Logf("Decode result: %v\n", target)
	}
}
