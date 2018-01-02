package distributor

import (
	"testing"

	"github.com/dearcode/sapper/distributor/config"
)

func TestProjectKey(t *testing.T) {
	config.Distributor.Server.SecretKey = "1qaz@WSX"
	ds := []struct {
		key string
		id  int64
	}{
		{"dhJgJns2tfBFvWVWUSGBfm1dsYVXAtTlye7csKmSgZo=", 1},
		{"+61FUC7/V/QxeZzpXV37e3jDOXEcAN3TXwFavJ1Ek9E=", 1234},
	}

	p := &project{}
	for _, data := range ds {
		p.ID = data.id
		if key := p.key(); key != data.key {
			t.Fatalf("invalid key:%v, expect:%v, id:%v", key, data.key, data.id)
		}
	}
}