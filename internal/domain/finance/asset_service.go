package finance

import (
	"context"
	"errors"
	"time"
)

type AssetRepository interface {
	Save(ctx context.Context, asset *Asset) error
	FindByID(ctx context.Context, id AssetID) (*Asset, error)
	FindByUserID(ctx context.Context, userID UserID, includeArchived bool) ([]*Asset, error)
}

type LiabilityRepository interface {
	Save(ctx context.Context, liability *Liability) error
	FindByID(ctx context.Context, id LiabilityID) (*Liability, error)
	FindByUserID(ctx context.Context, userID UserID, includeArchived bool) ([]*Liability, error)
}

type AssetService struct {
	assetRepo AssetRepository
}

func NewAssetService(assetRepo AssetRepository) *AssetService {
	return &AssetService{assetRepo: assetRepo}
}

func (s *AssetService) Create(
	ctx context.Context,
	userID UserID,
	name string,
	assetType AssetType,
	currencyID CurrencyID,
	currentValue float64,
	notes string,
	asOfDate *time.Time,
) (*Asset, error) {
	asset, err := NewAsset(userID, name, assetType, currencyID, currentValue, notes, asOfDate)
	if err != nil {
		return nil, err
	}
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *AssetService) List(ctx context.Context, userID UserID, includeArchived bool) ([]*Asset, error) {
	return s.assetRepo.FindByUserID(ctx, userID, includeArchived)
}

func (s *AssetService) GetForUser(ctx context.Context, userID UserID, id AssetID) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.UserID().Value() != userID.Value() {
		return nil, errors.New("asset not found")
	}
	return asset, nil
}

func (s *AssetService) Update(
	ctx context.Context,
	userID UserID,
	id AssetID,
	name string,
	assetType AssetType,
	currentValue float64,
	notes string,
	asOfDate *time.Time,
) (*Asset, error) {
	asset, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := asset.Update(name, assetType, currentValue, notes, asOfDate); err != nil {
		return nil, err
	}
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *AssetService) Archive(ctx context.Context, userID UserID, id AssetID) (*Asset, error) {
	asset, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	asset.Archive()
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *AssetService) Unarchive(ctx context.Context, userID UserID, id AssetID) (*Asset, error) {
	asset, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	asset.Unarchive()
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

type LiabilityService struct {
	liabilityRepo LiabilityRepository
}

func NewLiabilityService(liabilityRepo LiabilityRepository) *LiabilityService {
	return &LiabilityService{liabilityRepo: liabilityRepo}
}

func (s *LiabilityService) Create(
	ctx context.Context,
	userID UserID,
	name string,
	liabilityType LiabilityType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
) (*Liability, error) {
	liability, err := NewLiability(userID, name, liabilityType, currencyID, currentBalance, notes, asOfDate)
	if err != nil {
		return nil, err
	}
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return nil, err
	}
	return liability, nil
}

func (s *LiabilityService) List(ctx context.Context, userID UserID, includeArchived bool) ([]*Liability, error) {
	return s.liabilityRepo.FindByUserID(ctx, userID, includeArchived)
}

func (s *LiabilityService) GetForUser(ctx context.Context, userID UserID, id LiabilityID) (*Liability, error) {
	liability, err := s.liabilityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if liability.UserID().Value() != userID.Value() {
		return nil, errors.New("liability not found")
	}
	return liability, nil
}

func (s *LiabilityService) Update(
	ctx context.Context,
	userID UserID,
	id LiabilityID,
	name string,
	liabilityType LiabilityType,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
) (*Liability, error) {
	liability, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := liability.Update(name, liabilityType, currentBalance, notes, asOfDate); err != nil {
		return nil, err
	}
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return nil, err
	}
	return liability, nil
}

func (s *LiabilityService) Archive(ctx context.Context, userID UserID, id LiabilityID) (*Liability, error) {
	liability, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	liability.Archive()
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return nil, err
	}
	return liability, nil
}

func (s *LiabilityService) Unarchive(ctx context.Context, userID UserID, id LiabilityID) (*Liability, error) {
	liability, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	liability.Unarchive()
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return nil, err
	}
	return liability, nil
}
