package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
	"github.com/wyw/cry031-volunteer/internal/repository"
	"github.com/wyw/cry031-volunteer/internal/service"
)

type mutableClock struct {
	mu  sync.RWMutex
	now time.Time
}

func (c *mutableClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

func (c *mutableClock) Set(value time.Time) {
	c.mu.Lock()
	c.now = value
	c.mu.Unlock()
}

func fixture() (*Engine, *repository.MemoryStore, *mutableClock) {
	now := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	state := repository.DemoState(now)
	state.Memberships["member-new"] = domain.Membership{ID: "member-new", TeamID: "team-riverside", UserID: "u-new", Role: domain.RoleMember, Status: domain.MembershipActive, JoinedAt: now.Add(-time.Hour)}
	store := repository.NewMemoryStore(state)
	clock := &mutableClock{now: now}
	return NewEngine(store, clock, &service.LocalNotifier{}), store, clock
}

func captain() Actor   { return Actor{UserID: "u-captain", Role: domain.RoleCaptain} }
func volunteer() Actor { return Actor{UserID: "u-volunteer", Role: domain.RoleMember} }

func TestCancelledActivityCannotBeClaimedAndCancelsWaitlist(t *testing.T) {
	engine, store, clock := fixture()
	activity, err := engine.CreateActivity(context.Background(), captain(), RequestMeta{RequestID: "create"}, CreateActivityRequest{TeamID: "team-riverside", Title: "临时活动", StartAt: clock.Now().Add(24 * time.Hour), EndAt: clock.Now().Add(26 * time.Hour), Location: "服务站"})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.PublishActivity(context.Background(), captain(), RequestMeta{RequestID: "publish"}, activity.ID); err != nil {
		t.Fatal(err)
	}
	slot, err := engine.CreateSlot(context.Background(), captain(), RequestMeta{RequestID: "slot"}, CreateSlotRequest{ActivityID: activity.ID, Name: "引导", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ClaimSlot(context.Background(), volunteer(), RequestMeta{RequestID: "claim-1"}, slot.ID); err != nil {
		t.Fatal(err)
	}
	waiter, err := engine.ClaimSlot(context.Background(), Actor{UserID: "u-new", Role: domain.RoleMember}, RequestMeta{RequestID: "claim-2"}, slot.ID)
	if err != nil {
		t.Fatal(err)
	}
	if waiter.Status != domain.ClaimWaitlisted {
		t.Fatalf("waiter status = %s", waiter.Status)
	}
	if err := engine.CancelActivity(context.Background(), captain(), RequestMeta{RequestID: "cancel"}, activity.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ClaimSlot(context.Background(), Actor{UserID: "u-new", Role: domain.RoleMember}, RequestMeta{RequestID: "late"}, slot.ID); !errors.Is(err, domain.ErrActivityClosed) {
		t.Fatalf("late claim error = %v", err)
	}
	state, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, claim := range state.Claims {
		if claim.ActivityID == activity.ID && claim.Status != domain.ClaimCancelled {
			t.Fatalf("claim %s remained %s after cancellation", claim.ID, claim.Status)
		}
	}
}

func TestPromoteWaitlistDoesNotExceedCapacity(t *testing.T) {
	engine, store, clock := fixture()
	activity, err := engine.CreateActivity(context.Background(), captain(), RequestMeta{RequestID: "create"}, CreateActivityRequest{
		TeamID: "team-riverside", Title: "候补容量", StartAt: clock.Now().Add(24 * time.Hour), EndAt: clock.Now().Add(26 * time.Hour), Location: "广场",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.PublishActivity(context.Background(), captain(), RequestMeta{RequestID: "publish"}, activity.ID); err != nil {
		t.Fatal(err)
	}
	slot, err := engine.CreateSlot(context.Background(), captain(), RequestMeta{RequestID: "slot"}, CreateSlotRequest{ActivityID: activity.ID, Name: "岗位", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ClaimSlot(context.Background(), volunteer(), RequestMeta{RequestID: "claim-active"}, slot.ID); err != nil {
		t.Fatal(err)
	}
	waiter, err := engine.ClaimSlot(context.Background(), Actor{UserID: "u-new", Role: domain.RoleMember}, RequestMeta{RequestID: "claim-waiter"}, slot.ID)
	if err != nil {
		t.Fatal(err)
	}
	if waiter.Status != domain.ClaimWaitlisted {
		t.Fatalf("waiter status = %s", waiter.Status)
	}
	if err := engine.PromoteWaitlist(context.Background(), captain(), RequestMeta{RequestID: "promote-full"}, slot.ID); !errors.Is(err, domain.ErrCapacityExceeded) {
		t.Fatalf("promote full error = %v", err)
	}
	state, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Claims[waiter.ID].Status != domain.ClaimWaitlisted {
		t.Fatalf("waiter was promoted while full: %s", state.Claims[waiter.ID].Status)
	}
	if state.Slots[slot.ID].Status != domain.SlotFilled {
		t.Fatalf("slot status = %s", state.Slots[slot.ID].Status)
	}
}

func TestConcurrentClaimsRespectCapacity(t *testing.T) {
	engine, _, clock := fixture()
	activity, err := engine.CreateActivity(context.Background(), captain(), RequestMeta{RequestID: "create"}, CreateActivityRequest{TeamID: "team-riverside", Title: "并发活动", StartAt: clock.Now().Add(24 * time.Hour), EndAt: clock.Now().Add(26 * time.Hour), Location: "广场"})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.PublishActivity(context.Background(), captain(), RequestMeta{RequestID: "publish"}, activity.ID); err != nil {
		t.Fatal(err)
	}
	slot, err := engine.CreateSlot(context.Background(), captain(), RequestMeta{RequestID: "slot"}, CreateSlotRequest{ActivityID: activity.ID, Name: "岗位", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	actors := []Actor{volunteer(), {UserID: "u-new", Role: domain.RoleMember}}
	errs := make(chan error, len(actors))
	var wg sync.WaitGroup
	for _, actor := range actors {
		wg.Add(1)
		go func(actor Actor) {
			defer wg.Done()
			_, callErr := engine.ClaimSlot(context.Background(), actor, RequestMeta{RequestID: actor.UserID}, slot.ID)
			errs <- callErr
		}(actor)
	}
	wg.Wait()
	close(errs)
	success := 0
	conflict := 0
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrConflict) {
			conflict++
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("concurrent claim outcomes success=%d conflict=%d", success, conflict)
	}
}

func TestCaptainTransferKeepsSingleLeader(t *testing.T) {
	engine, store, _ := fixture()
	if err := engine.TransferCaptain(context.Background(), captain(), RequestMeta{RequestID: "transfer"}, "team-riverside", "u-new", "交接季度值班"); err != nil {
		t.Fatal(err)
	}
	state, _ := store.Snapshot(context.Background())
	team := state.Teams["team-riverside"]
	if team.CaptainID != "u-new" {
		t.Fatalf("captain = %s", team.CaptainID)
	}
	leaders := 0
	for _, member := range state.Memberships {
		if member.TeamID == team.ID && member.Role == domain.RoleCaptain {
			leaders++
		}
	}
	if leaders != 1 {
		t.Fatalf("leader count = %d", leaders)
	}
}

func TestMembershipPauseCanBeResumed(t *testing.T) {
	engine, store, _ := fixture()
	if err := engine.PauseMembership(context.Background(), volunteer(), RequestMeta{RequestID: "pause"}, "member-volunteer", true); err != nil {
		t.Fatal(err)
	}
	if err := engine.PauseMembership(context.Background(), volunteer(), RequestMeta{RequestID: "resume"}, "member-volunteer", false); err != nil {
		t.Fatal(err)
	}
	state, _ := store.Snapshot(context.Background())
	if state.Memberships["member-volunteer"].Status != domain.MembershipActive {
		t.Fatalf("membership status = %s", state.Memberships["member-volunteer"].Status)
	}
}

func TestCorrectionPreservesHistoryUntilApproved(t *testing.T) {
	engine, store, clock := fixture()
	activity := domain.Activity{ID: "activity-live", TeamID: "team-riverside", Title: "记录活动", Status: domain.ActivityPublished, StartAt: clock.Now().Add(-time.Hour), EndAt: clock.Now().Add(4 * time.Hour), Location: "服务站"}
	state, _ := store.Snapshot(context.Background())
	state.Activities[activity.ID] = activity
	state.Slots["slot-live"] = domain.RoleSlot{ID: "slot-live", ActivityID: activity.ID, Name: "岗位", Capacity: 1, Status: domain.SlotOpen}
	_ = store.Commit(context.Background(), state.Version, state, nil)
	claim, err := engine.ClaimSlot(context.Background(), volunteer(), RequestMeta{RequestID: "claim"}, "slot-live")
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.CheckIn(context.Background(), volunteer(), RequestMeta{RequestID: "check-in"}, claim.ID); err != nil {
		t.Fatal(err)
	}
	record, err := engine.StartService(context.Background(), volunteer(), RequestMeta{RequestID: "start"}, claim.ID)
	if err != nil {
		t.Fatal(err)
	}
	clock.Set(clock.Now().Add(2 * time.Hour))
	if _, err := engine.EndService(context.Background(), volunteer(), RequestMeta{RequestID: "end"}, record.ID, "现场照片已上传"); err != nil {
		t.Fatal(err)
	}
	if err := engine.ConfirmService(context.Background(), captain(), RequestMeta{RequestID: "confirm"}, record.ID); err != nil {
		t.Fatal(err)
	}
	start := clock.Now().Add(-3 * time.Hour)
	end := clock.Now()
	corrected, err := engine.RequestCorrection(context.Background(), volunteer(), RequestMeta{RequestID: "correct"}, record.ID, "签到时间修正", start, end)
	if err != nil {
		t.Fatal(err)
	}
	state, _ = store.Snapshot(context.Background())
	if state.Records[record.ID].Status != domain.ServiceConfirmed {
		t.Fatalf("original status before correction approval = %s", state.Records[record.ID].Status)
	}
	if err := engine.ConfirmService(context.Background(), captain(), RequestMeta{RequestID: "confirm-correction"}, corrected.ID); err != nil {
		t.Fatal(err)
	}
	csv, err := engine.ExportCertificate(context.Background(), volunteer().UserID)
	if err != nil {
		t.Fatal(err)
	}
	if string(csv) == "" || !contains(string(csv), ",180,confirmed") {
		t.Fatalf("certificate did not use corrected duration: %s", csv)
	}
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}

type failingStore struct{ state domain.State }

func (s *failingStore) Snapshot(context.Context) (domain.State, error) { return s.state, nil }
func (s *failingStore) Commit(context.Context, uint64, domain.State, []domain.AuditEvent) error {
	return errors.New("audit sink unavailable")
}

func TestTransactionFailureDoesNotMutateCallerSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	state := repository.DemoState(now)
	engine := NewEngine(&failingStore{state: state}, &mutableClock{now: now}, &service.LocalNotifier{})
	_, err := engine.CreateActivity(context.Background(), captain(), RequestMeta{RequestID: "failed"}, CreateActivityRequest{TeamID: "team-riverside", Title: "不会提交", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Location: "服务站"})
	if err == nil {
		t.Fatal("expected commit failure")
	}
}

func TestIdempotentActivityCreationReturnsSameResource(t *testing.T) {
	engine, store, clock := fixture()
	meta := RequestMeta{RequestID: "idem", IdempotencyKey: "activity-once"}
	request := CreateActivityRequest{TeamID: "team-riverside", Title: "幂等活动", StartAt: clock.Now().Add(time.Hour), EndAt: clock.Now().Add(2 * time.Hour), Location: "服务站"}
	first, err := engine.CreateActivity(context.Background(), captain(), meta, request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := engine.CreateActivity(context.Background(), captain(), meta, request)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotent IDs differ: %s vs %s", first.ID, second.ID)
	}
	state, _ := store.Snapshot(context.Background())
	count := 0
	for _, activity := range state.Activities {
		if activity.Title == request.Title {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("created %d resources", count)
	}
}

func TestRiskHandoffBlocksResolutionUntilAcknowledged(t *testing.T) {
	engine, store, _ := fixture()
	err := engine.ResolveRisk(context.Background(), captain(), RequestMeta{RequestID: "resolve-before"}, "risk-railing", "已经围挡")
	if !errors.Is(err, domain.ErrUnacknowledgedRisk) {
		t.Fatalf("resolve before handoff acknowledgement = %v", err)
	}
	if err := engine.AcknowledgeHandoff(context.Background(), volunteer(), RequestMeta{RequestID: "ack"}, "handoff-railing"); err != nil {
		t.Fatal(err)
	}
	if err := engine.ResolveRisk(context.Background(), volunteer(), RequestMeta{RequestID: "resolve-after"}, "risk-railing", "护栏已经修复并复查"); err != nil {
		t.Fatal(err)
	}
	state, _ := store.Snapshot(context.Background())
	if state.Risks["risk-railing"].Status != domain.RiskResolved {
		t.Fatalf("risk status = %s", state.Risks["risk-railing"].Status)
	}
}

func TestAnnouncementCreatesLocalNotice(t *testing.T) {
	engine, _, _ := fixture()
	if _, err := engine.CreateAnnouncement(context.Background(), captain(), RequestMeta{RequestID: "announce"}, "team-riverside", "集合调整", "集合地点改到东门"); err != nil {
		t.Fatal(err)
	}
	notices, err := engine.ListNotices(context.Background(), volunteer())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, notice := range notices {
		if notice.Kind == "announcement" && notice.Message == "集合调整" {
			found = true
		}
	}
	if !found {
		t.Fatal("announcement notice missing")
	}
}

func TestListRejectsUnknownSort(t *testing.T) {
	engine, _, _ := fixture()
	if _, err := engine.DiscoverTeams(context.Background(), TeamFilter{Sort: "drop_table"}); !errors.Is(err, domain.ErrInvalidFilter) {
		t.Fatalf("sort error = %v", err)
	}
}
