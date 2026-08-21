package claims

import (
	"context"
	"errors"
	"testing"
	"time"
)

type resolutionStoreStub struct{ claim Claim }

func (s *resolutionStoreStub) Load(context.Context, string) (Claim, error) { return s.claim, nil }
func (s *resolutionStoreStub) Save(_ context.Context, claim Claim) error {
	s.claim = claim
	return nil
}

type payoutGatewayStub struct{ err error }

func (s payoutGatewayStub) Pay(context.Context, Resolution) error { return s.err }

func TestSettlementFailureLeavesClaimOpenForRetry(t *testing.T) {
	now := time.Now().UTC()
	paymentErr := errors.New("payout gateway unavailable")
	store := &resolutionStoreStub{claim: Claim{ID: "claim-1", TenantID: "airline", ShipmentID: "shipment-1", FiledBy: "ops", Reason: "damaged cargo", Status: "open", FiledAt: now.Add(-time.Hour)}}
	resolver := Resolver{Store: store, Gateway: payoutGatewayStub{err: paymentErr}}
	err := resolver.Settle(context.Background(), Resolution{ClaimID: "claim-1", Amount: 25000, Reference: "settlement-1"}, now)
	if !errors.Is(err, paymentErr) {
		t.Fatalf("settlement error = %v", err)
	}
	if store.claim.Status != "open" || store.claim.ResolvedAt != nil {
		t.Fatalf("failed payout changed claim = %+v", store.claim)
	}
	if err := resolver.Settle(context.Background(), Resolution{ClaimID: "claim-1", Amount: 25000, Reference: "settlement-1"}, now); !errors.Is(err, paymentErr) {
		t.Fatalf("retry error = %v", err)
	}
}
