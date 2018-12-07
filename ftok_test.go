package ipc_test

import (
	"fmt"
	"testing"

	"github.com/siadat/ipc"
	"github.com/siadat/ipc/cgo_ftok"
)

func TestFtok(t *testing.T) {
	cases := []struct {
		path string
		id   uint64
	}{
		{"ftok.go", 50},
		{"ftok_test.go", 100},
	}

	for _, tt := range cases {
		key, err := ipc.Ftok(tt.path, tt.id)
		if err != nil {
			t.Fatal(err)
		}
		want, err := cgo_ftok.Ftok(tt.path, tt.id)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := fmt.Sprintf("0x%x", want), fmt.Sprintf("0x%x", key); got != want {
			t.Errorf("want %s, got %s", want, got)
		}
	}
}
