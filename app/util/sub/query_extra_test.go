// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package sub_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/app/util/sub"
)

func TestShareLinkQueryExtraFromStored_queryString(t *testing.T) {
	m := sub.ShareLinkQueryExtraFromStored("tx=10&foo=bar")
	if m["tx"] != "10" || m["foo"] != "bar" {
		t.Fatalf("got %#v", m)
	}
}

func TestShareLinkQueryExtraFromStored_json(t *testing.T) {
	m := sub.ShareLinkQueryExtraFromStored(`{"tx":"10","foo":"bar"}`)
	if m["tx"] != "10" || m["foo"] != "bar" {
		t.Fatalf("got %#v", m)
	}
}

func TestShareLinkQueryExtraFromStored_empty(t *testing.T) {
	if sub.ShareLinkQueryExtraFromStored("  ") != nil {
		t.Fatal("expected nil")
	}
}
