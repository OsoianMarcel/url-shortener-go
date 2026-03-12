package randlinkkey_test

import (
	"testing"

	"github.com/OsoianMarcel/url-shortener/pkg/randlinkkey"
)

func Test_GenLinkKeyLen(t *testing.T) {
	linkID := randlinkkey.GenLinkKey(6)

	if len(linkID) != 6 {
		t.Errorf("key length = %d; want 6", len(linkID))
	}
}

func Test_GenLinkKeyRandomness(t *testing.T) {
	key1 := randlinkkey.GenLinkKey(6)
	key2 := randlinkkey.GenLinkKey(6)

	if key1 == key2 {
		t.Errorf("two consecutive calls return the same key: %s", key1)
	}
}
