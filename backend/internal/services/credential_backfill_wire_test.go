package services

import (
	"context"
	"errors"
	"testing"
)

// Phase 10.3 — RED tests for the auto-backfill wire.
// Delegation must succeed, propagate errors, and refuse partial identifiers.

type fakeBackfillTrigger struct {
	called          bool
	lastTenantID    string
	lastPlatform    string
	lastStoreID     string
	returnErr       error
}

func (f *fakeBackfillTrigger) TriggerBackfill(_ context.Context, tenantID, platform, storeIdentifier string) error {
	f.called = true
	f.lastTenantID = tenantID
	f.lastPlatform = platform
	f.lastStoreID = storeIdentifier
	return f.returnErr
}

func TestInvokeBackfill_HappyDelegation(t *testing.T) {
	fake := &fakeBackfillTrigger{}
	err := invokeBackfill(context.Background(), fake, "yumna_bertigamart", "shopee", "530635055")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !fake.called {
		t.Fatalf("expected trigger to be called")
	}
	if fake.lastTenantID != "yumna_bertigamart" || fake.lastPlatform != "shopee" || fake.lastStoreID != "530635055" {
		t.Fatalf("unexpected args: tenant=%q platform=%q store=%q", fake.lastTenantID, fake.lastPlatform, fake.lastStoreID)
	}
}

func TestInvokeBackfill_PropagatesTriggerError(t *testing.T) {
	target := errors.New("shopee API 500")
	fake := &fakeBackfillTrigger{returnErr: target}
	err := invokeBackfill(context.Background(), fake, "t1", "shopee", "s1")
	if !errors.Is(err, target) {
		t.Fatalf("expected wrapped %v, got %v", target, err)
	}
}

func TestInvokeBackfill_NilTriggerReturnsSentinel(t *testing.T) {
	err := invokeBackfill(context.Background(), nil, "t1", "shopee", "s1")
	if !errors.Is(err, errBackfillNotWired) {
		t.Fatalf("expected errBackfillNotWired, got %v", err)
	}
}

func TestInvokeBackfill_RejectsMissingIdentifiers(t *testing.T) {
	fake := &fakeBackfillTrigger{}
	cases := []struct {
		name                   string
		tenant, plat, store    string
	}{
		{"missing tenant", "", "shopee", "s1"},
		{"missing platform", "t1", "", "s1"},
		{"missing store", "t1", "shopee", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := invokeBackfill(context.Background(), fake, tc.tenant, tc.plat, tc.store)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if fake.called {
				t.Fatalf("trigger should NOT be called when identifiers are missing (%s)", tc.name)
			}
		})
	}
}

func TestSetBackfillTrigger_WiresAndSupportsNil(t *testing.T) {
	svc := &CredentialApiService{}
	if svc.backfillTrigger != nil {
		t.Fatalf("fresh service should have nil trigger")
	}
	fake := &fakeBackfillTrigger{}
	svc.SetBackfillTrigger(fake)
	if svc.backfillTrigger == nil {
		t.Fatalf("SetBackfillTrigger should install the trigger")
	}
	svc.SetBackfillTrigger(nil)
	if svc.backfillTrigger != nil {
		t.Fatalf("SetBackfillTrigger(nil) should uninstall the trigger")
	}
}
