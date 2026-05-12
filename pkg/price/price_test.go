package price

import (
	"context"
	"testing"
)

func TestGetRejectsTooManyIDs(t *testing.T) {
	ids := make([]string, MaxIDsPerRequest+1)
	for i := range ids {
		ids[i] = "11111111111111111111111111111111"
	}
	_, err := (&Client{}).Get(context.Background(), GetRequest{IDs: ids})
	if err == nil {
		t.Fatal("expected too many ids error")
	}
}
