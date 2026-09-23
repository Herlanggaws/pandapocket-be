package ai

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/paas"
)

type stubStreamer struct {
	full      string
	err       error
	delay     time.Duration
	started   chan struct{}
	release   chan struct{}
}

func (s *stubStreamer) Configured() bool { return true }

func (s *stubStreamer) StreamChat(_ context.Context, _ []paas.Message, onDelta func(string) error) (string, int, int, error) {
	if s.started != nil {
		close(s.started)
	}
	if s.release != nil {
		<-s.release
	}
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	if s.err != nil {
		return "", 0, 0, s.err
	}
	if onDelta != nil && s.full != "" {
		_ = onDelta(s.full)
	}
	return s.full, 1, 1, nil
}

type memCredits struct {
	balance *domainAI.CreditBalance
}

func (m *memCredits) FindByUserID(_ context.Context, _ int) (*domainAI.CreditBalance, error) {
	if m.balance == nil {
		return nil, domainAI.ErrBalanceNotFound
	}
	cp := *m.balance
	return &cp, nil
}

func (m *memCredits) Save(_ context.Context, balance *domainAI.CreditBalance) error {
	cp := *balance
	m.balance = &cp
	return nil
}

func (m *memCredits) PurchaseExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *memCredits) RecordPurchase(_ context.Context, _ int, _, _ string, _ int) error {
	return nil
}

type memThreads struct {
	mu      sync.Mutex
	nextID  int
	threads map[int]*domainAI.Thread
	msgs    map[int][]domainAI.ThreadMessage
}

func newMemThreads() *memThreads {
	return &memThreads{
		nextID:  1,
		threads: map[int]*domainAI.Thread{},
		msgs:    map[int][]domainAI.ThreadMessage{},
	}
}

func (m *memThreads) ListByUserID(_ context.Context, userID int) ([]domainAI.Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domainAI.Thread, 0)
	for _, t := range m.threads {
		if t.UserID == userID {
			cp := *t
			out = append(out, cp)
		}
	}
	return out, nil
}

func (m *memThreads) Create(_ context.Context, userID int) (*domainAI.Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextID
	m.nextID++
	now := time.Now().UTC()
	t := &domainAI.Thread{
		ID:               id,
		UserID:           userID,
		GenerationStatus: domainAI.GenerationIdle,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.threads[id] = t
	m.msgs[id] = nil
	cp := *t
	return &cp, nil
}

func (m *memThreads) FindByIDForUser(_ context.Context, threadID, userID int) (*domainAI.Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok || t.UserID != userID {
		return nil, domainAI.ErrThreadNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *memThreads) Delete(_ context.Context, threadID, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok || t.UserID != userID {
		return domainAI.ErrThreadNotFound
	}
	delete(m.threads, threadID)
	delete(m.msgs, threadID)
	return nil
}

func (m *memThreads) UpdateTitle(_ context.Context, threadID, userID int, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok || t.UserID != userID {
		return domainAI.ErrThreadNotFound
	}
	t.Title = title
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *memThreads) SetTitleIfEmpty(_ context.Context, threadID int, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok {
		return domainAI.ErrThreadNotFound
	}
	if strings.TrimSpace(t.Title) == "" {
		t.Title = title
	}
	return nil
}

func (m *memThreads) TryBeginGeneration(_ context.Context, threadID, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok || t.UserID != userID {
		return domainAI.ErrThreadNotFound
	}
	if t.GenerationStatus == domainAI.GenerationPending {
		return domainAI.ErrTurnInProgress
	}
	now := time.Now().UTC()
	t.GenerationStatus = domainAI.GenerationPending
	t.GenerationStartedAt = &now
	return nil
}

func (m *memThreads) FinishGeneration(_ context.Context, threadID int, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[threadID]
	if !ok {
		return domainAI.ErrThreadNotFound
	}
	t.GenerationStatus = domainAI.GenerationIdle
	t.GenerationStartedAt = nil
	return nil
}

func (m *memThreads) ListMessages(_ context.Context, threadID int) ([]domainAI.ThreadMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	src := m.msgs[threadID]
	out := make([]domainAI.ThreadMessage, len(src))
	copy(out, src)
	return out, nil
}

func (m *memThreads) AppendMessage(_ context.Context, threadID int, role, content string, _, _ int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msgs[threadID] = append(m.msgs[threadID], domainAI.ThreadMessage{Role: role, Content: content})
	if t, ok := m.threads[threadID]; ok {
		t.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (m *memThreads) ClearMessages(_ context.Context, threadID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msgs[threadID] = nil
	return nil
}

func (m *memThreads) TrimOldest(_ context.Context, _, _ int) error { return nil }

func (m *memThreads) FirstUserMessageContent(_ context.Context, threadID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, msg := range m.msgs[threadID] {
		if msg.Role == domainAI.RoleUser && strings.TrimSpace(msg.Content) != "" {
			return msg.Content, nil
		}
	}
	return "", nil
}

type alwaysPro struct{}

func (alwaysPro) IsPro(_ context.Context, _ int) (bool, error) { return true, nil }

type emptySubs struct{}

func (emptySubs) Save(_ context.Context, _ *domainBilling.Subscription) error { return nil }
func (emptySubs) FindByUserID(_ context.Context, _ int) (*domainBilling.Subscription, error) {
	return nil, domainBilling.ErrNotFound
}
func (emptySubs) ListAll(_ context.Context) ([]*domainBilling.Subscription, error) {
	return nil, nil
}

func newTestChat(streamer *stubStreamer) (*AdvisorChatUseCase, *memCredits, *memThreads, int) {
	bal := domainAI.NewCreditBalance(1, false)
	creditsRepo := &memCredits{balance: bal}
	svc := NewCreditService(creditsRepo, emptySubs{})
	threads := newMemThreads()
	t, _ := threads.Create(context.Background(), 1)
	uc := NewAdvisorChatUseCase(svc, threads, alwaysPro{}, streamer, nil, nil)
	return uc, creditsRepo, threads, t.ID
}

func drain(ch <-chan StreamEvent) *StreamEvent {
	var last *StreamEvent
	for ev := range ch {
		cp := ev
		last = &cp
		if ev.Kind == StreamDone || ev.Kind == StreamError {
			break
		}
	}
	return last
}

func TestAdvisorChatSpendsOnlyAfterSuccess(t *testing.T) {
	uc, repo, _, threadID := newTestChat(&stubStreamer{full: "Halo, ini jawaban."})
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "Apa kabar?")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer unsub()
	last := drain(ch)
	if last == nil || last.Kind != StreamDone {
		t.Fatalf("want done, got %#v", last)
	}
	if last.Credits == nil || last.Credits.Available != domainAI.IncludedGrant()-1 {
		t.Fatalf("available=%v", last.Credits)
	}
	if repo.balance.Available() != domainAI.IncludedGrant()-1 {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}

func TestAdvisorChatNoSpendOnUpstreamFailure(t *testing.T) {
	uc, repo, _, threadID := newTestChat(&stubStreamer{err: errors.New("provider down")})
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "Apa kabar?")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer unsub()
	last := drain(ch)
	if last == nil || last.Kind != StreamError {
		t.Fatalf("want error, got %#v", last)
	}
	if repo.balance.Available() != domainAI.IncludedGrant() {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}

func TestAdvisorChatNoSpendOnEmptyReply(t *testing.T) {
	uc, repo, _, threadID := newTestChat(&stubStreamer{full: "   "})
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "Apa kabar?")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer unsub()
	_ = drain(ch)
	if repo.balance.Available() != domainAI.IncludedGrant() {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}

func TestAdvisorChatRejectsSecondWhilePending(t *testing.T) {
	started := make(chan struct{})
	cont := make(chan struct{})
	uc, _, threads, threadID := newTestChat(&stubStreamer{full: "ok", started: started, release: cont})
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "Pertama")
	if err != nil {
		t.Fatalf("start1: %v", err)
	}
	defer unsub()
	<-started
	_, _, err = uc.Start(context.Background(), 1, threadID, "Kedua")
	if !errors.Is(err, domainAI.ErrTurnInProgress) {
		t.Fatalf("want turn in progress, got %v", err)
	}
	close(cont)
	_ = drain(ch)
	t2, _ := threads.FindByIDForUser(context.Background(), threadID, 1)
	if t2.GenerationStatus != domainAI.GenerationIdle {
		t.Fatalf("status=%s", t2.GenerationStatus)
	}
}

func TestThreadListDerivesTitleFromFirstMessage(t *testing.T) {
	bal := domainAI.NewCreditBalance(1, false)
	credits := NewCreditService(&memCredits{balance: bal}, emptySubs{})
	threads := newMemThreads()
	uc := NewThreadUseCases(credits, threads, alwaysPro{})

	created, err := uc.Create(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := threads.AppendMessage(context.Background(), created.ID, domainAI.RoleUser, "Kenapa budget saya jebol bulan ini?", 0, 0); err != nil {
		t.Fatal(err)
	}
	list, err := uc.List(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
	want := "Kenapa budget saya jebol bulan ini?"
	if list[0].Title != want {
		t.Fatalf("title=%q want %q", list[0].Title, want)
	}
	stored, err := threads.FindByIDForUser(context.Background(), created.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Title != want {
		t.Fatalf("persisted title=%q", stored.Title)
	}
}

func TestThreadListCreateDelete(t *testing.T) {
	bal := domainAI.NewCreditBalance(1, false)
	credits := NewCreditService(&memCredits{balance: bal}, emptySubs{})
	threads := newMemThreads()
	uc := NewThreadUseCases(credits, threads, alwaysPro{})

	a, err := uc.Create(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := uc.Create(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	list, err := uc.List(context.Background(), 1)
	if err != nil || len(list) != 2 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if err := uc.Delete(context.Background(), 1, a.ID); err != nil {
		t.Fatal(err)
	}
	got, err := uc.Get(context.Background(), 1, b.ID)
	if err != nil || got.ID != b.ID {
		t.Fatalf("get=%v err=%v", got, err)
	}
	_, err = uc.Get(context.Background(), 1, a.ID)
	if !errors.Is(err, domainAI.ErrThreadNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
}
