package identity

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

type memUserRepo struct {
	mu      sync.Mutex
	users   map[int]*domainIdentity.User
	emails  map[string]int
	deleted map[int]time.Time
	nextID  int
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		users:   map[int]*domainIdentity.User{},
		emails:  map[string]int{},
		deleted: map[int]time.Time{},
		nextID:  1,
	}
}

func (r *memUserRepo) Save(ctx context.Context, user *domainIdentity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := user.ID().Value()
	if id == 0 {
		id = r.nextID
		r.nextID++
		user.AssignID(domainIdentity.NewUserID(id))
	}
	r.users[id] = user
	r.emails[user.Email().Value()] = id
	return nil
}

func (r *memUserRepo) Update(ctx context.Context, user *domainIdentity.User) error {
	return r.Save(ctx, user)
}

func (r *memUserRepo) FindByID(ctx context.Context, id domainIdentity.UserID) (*domainIdentity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, deleted := r.deleted[id.Value()]; deleted {
		return nil, errors.New("not found")
	}
	user, ok := r.users[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (r *memUserRepo) FindByEmail(ctx context.Context, email domainIdentity.Email) (*domainIdentity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.emails[email.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	if _, deleted := r.deleted[id]; deleted {
		return nil, errors.New("not found")
	}
	return r.users[id], nil
}

func (r *memUserRepo) FindAll(ctx context.Context) ([]*domainIdentity.User, error) {
	return nil, nil
}

func (r *memUserRepo) Delete(ctx context.Context, id domainIdentity.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id.Value())
	delete(r.deleted, id.Value())
	return nil
}

func (r *memUserRepo) ExistsByEmail(ctx context.Context, email domainIdentity.Email) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.emails[email.Value()]
	if !ok {
		return false, nil
	}
	_, deleted := r.deleted[id]
	return !deleted, nil
}

func (r *memUserRepo) ExistsActive(ctx context.Context, id domainIdentity.UserID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, deleted := r.deleted[id.Value()]; deleted {
		return false, nil
	}
	_, ok := r.users[id.Value()]
	return ok, nil
}

func (r *memUserRepo) SoftDelete(ctx context.Context, id domainIdentity.UserID, mangledEmail string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[id.Value()]
	if !ok {
		return errors.New("not found")
	}
	if _, deleted := r.deleted[id.Value()]; deleted {
		return errors.New("not found")
	}
	for email, uid := range r.emails {
		if uid == id.Value() {
			delete(r.emails, email)
		}
	}
	_ = user.ChangeEmail(mustEmail(mangledEmail))
	user.MarkDeleted(deletedAt)
	r.emails[mangledEmail] = id.Value()
	r.deleted[id.Value()] = deletedAt
	return nil
}

func (r *memUserRepo) ListDueForPurge(ctx context.Context, before time.Time) ([]domainIdentity.UserID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ids []domainIdentity.UserID
	for id, deletedAt := range r.deleted {
		if !deletedAt.After(before) {
			ids = append(ids, domainIdentity.NewUserID(id))
		}
	}
	return ids, nil
}

func mustEmail(value string) domainIdentity.Email {
	email, err := domainIdentity.NewEmail(value)
	if err != nil {
		panic(err)
	}
	return email
}

func seedUser(t *testing.T, repo *memUserRepo, email, password string) *domainIdentity.User {
	t.Helper()
	emailVO, err := domainIdentity.NewEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := domainIdentity.NewPasswordHashFromPlain(password)
	if err != nil {
		t.Fatal(err)
	}
	role, err := domainIdentity.NewRole("user")
	if err != nil {
		t.Fatal(err)
	}
	user := domainIdentity.NewUser(domainIdentity.UserID{}, emailVO, hash, role)
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	return user
}

type memTokenRevoker struct {
	mu           sync.Mutex
	revokedUsers []int
	err          error
}

func (r *memTokenRevoker) RevokeAllForUser(ctx context.Context, userID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.revokedUsers = append(r.revokedUsers, userID)
	return nil
}

type memHardDeleter struct {
	mu           sync.Mutex
	deletedUsers []uint
	err          error
}

func (d *memHardDeleter) HardDeleteAccount(ctx context.Context, userID uint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return d.err
	}
	d.deletedUsers = append(d.deletedUsers, userID)
	return nil
}

func TestDeleteAccountRejectsInvalidPassword(t *testing.T) {
	repo := newMemUserRepo()
	user := seedUser(t, repo, "user@example.com", "correct-password")
	revoker := &memTokenRevoker{}
	uc := NewDeleteAccountUseCase(repo, revoker)

	_, err := uc.Execute(context.Background(), DeleteAccountRequest{
		UserID:   user.ID().Value(),
		Password: "wrong-password",
	})
	if err == nil || err.Error() != "invalid password" {
		t.Fatalf("expected invalid password, got %v", err)
	}
	if len(revoker.revokedUsers) != 0 {
		t.Fatalf("expected no revoke, got %v", revoker.revokedUsers)
	}
}

func TestDeleteAccountSoftDeletesAndRevokes(t *testing.T) {
	repo := newMemUserRepo()
	user := seedUser(t, repo, "user@example.com", "correct-password")
	revoker := &memTokenRevoker{}
	uc := NewDeleteAccountUseCase(repo, revoker)
	fixedNow := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }

	resp, err := uc.Execute(context.Background(), DeleteAccountRequest{
		UserID:   user.ID().Value(),
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.ScheduledPurgeAt != fixedNow.Add(AccountDeletionRetention) {
		t.Fatalf("unexpected purge at %v", resp.ScheduledPurgeAt)
	}

	active, err := repo.ExistsActive(context.Background(), user.ID())
	if err != nil {
		t.Fatal(err)
	}
	if active {
		t.Fatal("expected user inactive after soft delete")
	}
	if len(revoker.revokedUsers) != 1 || revoker.revokedUsers[0] != user.ID().Value() {
		t.Fatalf("expected revoke for user, got %v", revoker.revokedUsers)
	}

	exists, err := repo.ExistsByEmail(context.Background(), mustEmail("user@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("original email should be free after soft delete")
	}
}

func TestPurgeDeletedAccountsOnlyDue(t *testing.T) {
	repo := newMemUserRepo()
	dueUser := seedUser(t, repo, "due@example.com", "password123")
	freshUser := seedUser(t, repo, "fresh@example.com", "password123")

	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	if err := repo.SoftDelete(context.Background(), dueUser.ID(), "deleted+due@deleted.local", now.Add(-15*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SoftDelete(context.Background(), freshUser.ID(), "deleted+fresh@deleted.local", now.Add(-2*24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	deleter := &memHardDeleter{}
	uc := NewPurgeDeletedAccountsUseCase(repo, deleter)
	uc.now = func() time.Time { return now }

	purged, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if purged != 1 {
		t.Fatalf("expected 1 purged, got %d", purged)
	}
	if len(deleter.deletedUsers) != 1 || deleter.deletedUsers[0] != uint(dueUser.ID().Value()) {
		t.Fatalf("expected due user purged, got %v", deleter.deletedUsers)
	}
}
