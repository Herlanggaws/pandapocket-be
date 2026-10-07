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

type LiabilityPaymentRepository interface {
	Save(ctx context.Context, payment *LiabilityPayment) error
	FindByLiabilityID(ctx context.Context, liabilityID LiabilityID) ([]*LiabilityPayment, error)
	FindByExpenseID(ctx context.Context, expenseID int) (*LiabilityPayment, error)
	Delete(ctx context.Context, id LiabilityPaymentID) error
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
	paymentRepo   LiabilityPaymentRepository
}

func NewLiabilityService(liabilityRepo LiabilityRepository, paymentRepo LiabilityPaymentRepository) *LiabilityService {
	return &LiabilityService{liabilityRepo: liabilityRepo, paymentRepo: paymentRepo}
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
	debt LiabilityDebtDetails,
) (*Liability, error) {
	liability, err := NewLiability(userID, name, liabilityType, currencyID, currentBalance, notes, asOfDate, debt)
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
	debt LiabilityDebtDetails,
) (*Liability, error) {
	liability, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := liability.Update(name, liabilityType, currentBalance, notes, asOfDate, debt); err != nil {
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

func (s *LiabilityService) RecordPayment(
	ctx context.Context,
	userID UserID,
	id LiabilityID,
	amount float64,
	paidAt time.Time,
	expenseID *int,
	note string,
) (*Liability, *LiabilityPayment, error) {
	liability, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if err := liability.ApplyPayment(amount); err != nil {
		return nil, nil, err
	}
	payment, err := NewLiabilityPayment(id, userID, amount, paidAt, expenseID, note)
	if err != nil {
		return nil, nil, err
	}
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return nil, nil, err
	}
	if s.paymentRepo != nil {
		if err := s.paymentRepo.Save(ctx, payment); err != nil {
			return nil, nil, err
		}
	}
	return liability, payment, nil
}

func (s *LiabilityService) ListPayments(ctx context.Context, userID UserID, id LiabilityID) ([]*LiabilityPayment, error) {
	if _, err := s.GetForUser(ctx, userID, id); err != nil {
		return nil, err
	}
	if s.paymentRepo == nil {
		return []*LiabilityPayment{}, nil
	}
	return s.paymentRepo.FindByLiabilityID(ctx, id)
}

func (s *LiabilityService) ReversePaymentByExpenseID(ctx context.Context, userID UserID, expenseID int) error {
	if s.paymentRepo == nil {
		return nil
	}
	payment, err := s.paymentRepo.FindByExpenseID(ctx, expenseID)
	if err != nil {
		return err
	}
	if payment == nil {
		return nil
	}
	if payment.UserID().Value() != userID.Value() {
		return errors.New("liability payment not found")
	}

	liability, err := s.GetForUser(ctx, userID, payment.LiabilityID())
	if err != nil {
		return err
	}
	if err := liability.ReversePayment(payment.Amount()); err != nil {
		return err
	}
	if err := s.liabilityRepo.Save(ctx, liability); err != nil {
		return err
	}
	return s.paymentRepo.Delete(ctx, payment.ID())
}

type ReceivableRepository interface {
	Save(ctx context.Context, receivable *Receivable) error
	FindByID(ctx context.Context, id ReceivableID) (*Receivable, error)
	FindByUserID(ctx context.Context, userID UserID, includeArchived bool) ([]*Receivable, error)
	FindByCreateExpenseID(ctx context.Context, expenseID int) (*Receivable, error)
}

type ReceivableCollectionRepository interface {
	Save(ctx context.Context, collection *ReceivableCollection) error
	FindByReceivableID(ctx context.Context, receivableID ReceivableID) ([]*ReceivableCollection, error)
	FindByIncomeID(ctx context.Context, incomeID int) (*ReceivableCollection, error)
	Delete(ctx context.Context, id ReceivableCollectionID) error
}

type ReceivableService struct {
	receivableRepo ReceivableRepository
	collectionRepo ReceivableCollectionRepository
}

func NewReceivableService(
	receivableRepo ReceivableRepository,
	collectionRepo ReceivableCollectionRepository,
) *ReceivableService {
	return &ReceivableService{receivableRepo: receivableRepo, collectionRepo: collectionRepo}
}

func (s *ReceivableService) Create(
	ctx context.Context,
	userID UserID,
	name string,
	receivableType ReceivableType,
	currencyID CurrencyID,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	details ReceivableDetails,
) (*Receivable, error) {
	receivable, err := NewReceivable(userID, name, receivableType, currencyID, currentBalance, notes, asOfDate, details)
	if err != nil {
		return nil, err
	}
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, err
	}
	return receivable, nil
}

func (s *ReceivableService) List(ctx context.Context, userID UserID, includeArchived bool) ([]*Receivable, error) {
	return s.receivableRepo.FindByUserID(ctx, userID, includeArchived)
}

func (s *ReceivableService) GetForUser(ctx context.Context, userID UserID, id ReceivableID) (*Receivable, error) {
	receivable, err := s.receivableRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if receivable.UserID().Value() != userID.Value() {
		return nil, errors.New("receivable not found")
	}
	return receivable, nil
}

func (s *ReceivableService) Update(
	ctx context.Context,
	userID UserID,
	id ReceivableID,
	name string,
	receivableType ReceivableType,
	currentBalance float64,
	notes string,
	asOfDate *time.Time,
	details ReceivableDetails,
) (*Receivable, error) {
	receivable, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := receivable.Update(name, receivableType, currentBalance, notes, asOfDate, details); err != nil {
		return nil, err
	}
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, err
	}
	return receivable, nil
}

func (s *ReceivableService) Archive(ctx context.Context, userID UserID, id ReceivableID) (*Receivable, error) {
	receivable, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	receivable.Archive()
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, err
	}
	return receivable, nil
}

func (s *ReceivableService) Unarchive(ctx context.Context, userID UserID, id ReceivableID) (*Receivable, error) {
	receivable, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	receivable.Unarchive()
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, err
	}
	return receivable, nil
}

func (s *ReceivableService) AttachCreateExpenseID(
	ctx context.Context,
	userID UserID,
	id ReceivableID,
	expenseID int,
) (*Receivable, error) {
	receivable, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	receivable.SetCreateExpenseID(&expenseID)
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, err
	}
	return receivable, nil
}

func (s *ReceivableService) RecordCollection(
	ctx context.Context,
	userID UserID,
	id ReceivableID,
	amount float64,
	collectedAt time.Time,
	incomeID *int,
	note string,
) (*Receivable, *ReceivableCollection, error) {
	receivable, err := s.GetForUser(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if err := receivable.ApplyCollection(amount); err != nil {
		return nil, nil, err
	}
	collection, err := NewReceivableCollection(id, userID, amount, collectedAt, incomeID, note)
	if err != nil {
		return nil, nil, err
	}
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return nil, nil, err
	}
	if s.collectionRepo != nil {
		if err := s.collectionRepo.Save(ctx, collection); err != nil {
			return nil, nil, err
		}
	}
	return receivable, collection, nil
}

func (s *ReceivableService) ListCollections(ctx context.Context, userID UserID, id ReceivableID) ([]*ReceivableCollection, error) {
	if _, err := s.GetForUser(ctx, userID, id); err != nil {
		return nil, err
	}
	if s.collectionRepo == nil {
		return []*ReceivableCollection{}, nil
	}
	return s.collectionRepo.FindByReceivableID(ctx, id)
}

func (s *ReceivableService) ReverseCollectionByIncomeID(ctx context.Context, userID UserID, incomeID int) error {
	if s.collectionRepo == nil {
		return nil
	}
	collection, err := s.collectionRepo.FindByIncomeID(ctx, incomeID)
	if err != nil {
		return err
	}
	if collection == nil {
		return nil
	}
	if collection.UserID().Value() != userID.Value() {
		return errors.New("receivable collection not found")
	}

	receivable, err := s.GetForUser(ctx, userID, collection.ReceivableID())
	if err != nil {
		return err
	}
	if err := receivable.ReverseCollection(collection.Amount()); err != nil {
		return err
	}
	if err := s.receivableRepo.Save(ctx, receivable); err != nil {
		return err
	}
	return s.collectionRepo.Delete(ctx, collection.ID())
}

func (s *ReceivableService) ReverseCreateExpenseByExpenseID(ctx context.Context, userID UserID, expenseID int) error {
	receivable, err := s.receivableRepo.FindByCreateExpenseID(ctx, expenseID)
	if err != nil {
		return err
	}
	if receivable == nil {
		return nil
	}
	if receivable.UserID().Value() != userID.Value() {
		return errors.New("receivable not found")
	}
	receivable.Archive()
	receivable.SetCreateExpenseID(nil)
	return s.receivableRepo.Save(ctx, receivable)
}
