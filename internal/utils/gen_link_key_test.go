package util_test

import (
	"testing"

	util "github.com/OsoianMarcel/url-shortener/internal/utils"
)

func Test_GenLinkKeyLen(t *testing.T) {
	linkID := util.GenLinkKey()

	if len(linkID) != 6 {
		t.Errorf("key length = %d; want 6", len(linkID))
	}
}

func Test_GenLinkKeyRandomness(t *testing.T) {
	key1 := util.GenLinkKey()
	key2 := util.GenLinkKey()

	if key1 == key2 {
		t.Errorf("two consecutive calls return the same key: %s", key1)
	}
}
