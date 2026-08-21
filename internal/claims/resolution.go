package claims

import (
	"context"
	"time"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
)

type ResolutionStore interface {
	Load(context.Context, string) (Claim, error)
	Save(context.Context, Claim) error
}

type PayoutGateway interface {
	Pay(context.Context, Resolution) error
}

type Resolver struct {
	Store   ResolutionStore
	Gateway PayoutGateway
}

func (r Resolver) Settle(ctx context.Context, resolution Resolution, at time.Time) error {
	if r.Store == nil || r.Gateway == nil || resolution.ClaimID == "" || resolution.Amount <= 0 || resolution.Reference == "" {
		return domain.ErrInvalid
	}
	claim, err := r.Store.Load(ctx, resolution.ClaimID)
	if err != nil {
		return err
	}
	resolved, err := Resolve(claim, at)
	if err != nil {
		return err
	}
	if err := r.Store.Save(ctx, resolved); err != nil {
		return err
	}
	if err := r.Gateway.Pay(ctx, resolution); err != nil {
		return err
	}
	return nil
}
