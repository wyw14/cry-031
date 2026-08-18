package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type UserRole string

const (
	RoleResident UserRole = "resident"
	RoleMember   UserRole = "member"
	RoleCaptain  UserRole = "captain"
	RoleAdmin    UserRole = "admin"
)

type TeamStatus string

const (
	TeamRecruiting TeamStatus = "recruiting"
	TeamPaused     TeamStatus = "paused"
	TeamClosed     TeamStatus = "closed"
)

type MembershipStatus string

const (
	MembershipPending MembershipStatus = "pending"
	MembershipActive  MembershipStatus = "active"
	MembershipPaused  MembershipStatus = "paused"
	MembershipExited  MembershipStatus = "exited"
)

type ActivityStatus string

const (
	ActivityDraft     ActivityStatus = "draft"
	ActivityPublished ActivityStatus = "published"
	ActivityCancelled ActivityStatus = "cancelled"
	ActivityCompleted ActivityStatus = "completed"
)

type SlotStatus string

const (
	SlotOpen      SlotStatus = "open"
	SlotFilled    SlotStatus = "filled"
	SlotCancelled SlotStatus = "cancelled"
)

type ClaimStatus string

const (
	ClaimActive     ClaimStatus = "active"
	ClaimWaitlisted ClaimStatus = "waitlisted"
	ClaimCancelled  ClaimStatus = "cancelled"
	ClaimCheckedIn  ClaimStatus = "checked_in"
	ClaimNoShow     ClaimStatus = "no_show"
)

type ServiceStatus string

const (
	ServiceStarted        ServiceStatus = "started"
	ServiceEnded          ServiceStatus = "ended"
	ServicePendingReview  ServiceStatus = "pending_review"
	ServiceConfirmed      ServiceStatus = "confirmed"
	ServiceCorrectionOpen ServiceStatus = "correction_open"
)

type RiskStatus string

const (
	RiskOpen       RiskStatus = "open"
	RiskInProgress RiskStatus = "in_progress"
	RiskResolved   RiskStatus = "resolved"
)

type HandoffStatus string

const (
	HandoffPending      HandoffStatus = "pending"
	HandoffAcknowledged HandoffStatus = "acknowledged"
	HandoffOverdue      HandoffStatus = "overdue"
	HandoffCompleted    HandoffStatus = "completed"
)

type Community struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ServiceArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID             string     `json:"id"`
	CommunityID    string     `json:"community_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Status         TeamStatus `json:"status"`
	Visibility     string     `json:"visibility"`
	ServiceAreaIDs []string   `json:"service_area_ids"`
	InviteCode     string     `json:"invite_code"`
	CaptainID      string     `json:"captain_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

type MembershipEvent struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	ActorID   string    `json:"actor_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type Membership struct {
	ID       string            `json:"id"`
	TeamID   string            `json:"team_id"`
	UserID   string            `json:"user_id"`
	Role     UserRole          `json:"role"`
	Status   MembershipStatus  `json:"status"`
	JoinedAt time.Time         `json:"joined_at"`
	ExitedAt *time.Time        `json:"exited_at,omitempty"`
	Events   []MembershipEvent `json:"events"`
}

type User struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Phone       string    `json:"phone"`
	Role        UserRole  `json:"role"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Activity struct {
	ID            string         `json:"id"`
	TeamID        string         `json:"team_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Status        ActivityStatus `json:"status"`
	StartAt       time.Time      `json:"start_at"`
	EndAt         time.Time      `json:"end_at"`
	Location      string         `json:"location"`
	MaterialNeeds []string       `json:"material_needs"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	CancelledAt   *time.Time     `json:"cancelled_at,omitempty"`
}

type RoleSlot struct {
	ID          string     `json:"id"`
	ActivityID  string     `json:"activity_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Capacity    int        `json:"capacity"`
	Status      SlotStatus `json:"status"`
}

type Claim struct {
	ID         string      `json:"id"`
	SlotID     string      `json:"slot_id"`
	ActivityID string      `json:"activity_id"`
	UserID     string      `json:"user_id"`
	Status     ClaimStatus `json:"status"`
	ClaimedAt  time.Time   `json:"claimed_at"`
	CheckedAt  *time.Time  `json:"checked_at,omitempty"`
}

type ServiceRecord struct {
	ID             string        `json:"id"`
	SupersedesID   string        `json:"supersedes_id,omitempty"`
	ClaimID        string        `json:"claim_id"`
	ActivityID     string        `json:"activity_id"`
	UserID         string        `json:"user_id"`
	Version        int           `json:"version"`
	Status         ServiceStatus `json:"status"`
	StartedAt      *time.Time    `json:"started_at,omitempty"`
	EndedAt        *time.Time    `json:"ended_at,omitempty"`
	DurationMins   int           `json:"duration_minutes"`
	PhotoNote      string        `json:"photo_note"`
	ReviewerID     string        `json:"reviewer_id"`
	CorrectionNote string        `json:"correction_note"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type RiskItem struct {
	ID          string     `json:"id"`
	ActivityID  string     `json:"activity_id"`
	TeamID      string     `json:"team_id"`
	Title       string     `json:"title"`
	Severity    string     `json:"severity"`
	Description string     `json:"description"`
	OwnerID     string     `json:"owner_id"`
	Status      RiskStatus `json:"status"`
	DueAt       time.Time  `json:"due_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

type FollowUp struct {
	ID        string    `json:"id"`
	RiskID    string    `json:"risk_id"`
	Title     string    `json:"title"`
	OwnerID   string    `json:"owner_id"`
	DueAt     time.Time `json:"due_at"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type Handoff struct {
	ID         string        `json:"id"`
	ActivityID string        `json:"activity_id"`
	RiskID     string        `json:"risk_id"`
	FromUserID string        `json:"from_user_id"`
	ToUserID   string        `json:"to_user_id"`
	Status     HandoffStatus `json:"status"`
	Note       string        `json:"note"`
	DueAt      time.Time     `json:"due_at"`
	AckAt      *time.Time    `json:"ack_at,omitempty"`
}

type Announcement struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	AuthorID  string    `json:"author_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Notice struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditEvent struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	RequestID  string    `json:"request_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// State is the aggregate snapshot persisted by the repository. The store uses optimistic versioning.
type State struct {
	Version       uint64                   `json:"version"`
	Users         map[string]User          `json:"users"`
	Communities   map[string]Community     `json:"communities"`
	Areas         map[string]ServiceArea   `json:"areas"`
	Teams         map[string]Team          `json:"teams"`
	Memberships   map[string]Membership    `json:"memberships"`
	Activities    map[string]Activity      `json:"activities"`
	Slots         map[string]RoleSlot      `json:"slots"`
	Claims        map[string]Claim         `json:"claims"`
	Records       map[string]ServiceRecord `json:"records"`
	Risks         map[string]RiskItem      `json:"risks"`
	FollowUps     map[string]FollowUp      `json:"follow_ups"`
	Handoffs      map[string]Handoff       `json:"handoffs"`
	Announcements map[string]Announcement  `json:"announcements"`
	Notices       map[string]Notice        `json:"notices"`
	Audit         []AuditEvent             `json:"audit"`
	Idempotency   map[string]string        `json:"idempotency"`
}

func NewState() State {
	return State{
		Users: map[string]User{}, Communities: map[string]Community{}, Areas: map[string]ServiceArea{},
		Teams: map[string]Team{}, Memberships: map[string]Membership{}, Activities: map[string]Activity{},
		Slots: map[string]RoleSlot{}, Claims: map[string]Claim{}, Records: map[string]ServiceRecord{},
		Risks: map[string]RiskItem{}, FollowUps: map[string]FollowUp{}, Handoffs: map[string]Handoff{},
		Announcements: map[string]Announcement{}, Notices: map[string]Notice{}, Audit: []AuditEvent{}, Idempotency: map[string]string{},
	}
}

func (s State) TeamMembers(teamID string) []Membership {
	result := make([]Membership, 0)
	for _, member := range s.Memberships {
		if member.TeamID == teamID {
			result = append(result, member)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].JoinedAt.Before(result[j].JoinedAt) })
	return result
}

func (a Activity) CanClaim(now time.Time) error {
	if a.Status != ActivityPublished || now.After(a.EndAt) || a.CancelledAt != nil {
		return ErrActivityClosed
	}
	return nil
}

func (a Activity) WindowOverlaps(other Activity) bool {
	return a.StartAt.Before(other.EndAt) && other.StartAt.Before(a.EndAt)
}

func (r ServiceRecord) Duration() (int, error) {
	if r.StartedAt == nil || r.EndedAt == nil || !r.EndedAt.After(*r.StartedAt) {
		return 0, ErrInvalidDuration
	}
	minutes := int(r.EndedAt.Sub(*r.StartedAt).Minutes())
	if minutes < 1 || minutes > 24*60 {
		return 0, ErrInvalidDuration
	}
	return minutes, nil
}

func (r ServiceRecord) PublicSummary() map[string]any {
	return map[string]any{"id": r.ID, "user_id": r.UserID, "activity_id": r.ActivityID, "version": r.Version, "status": r.Status, "duration_minutes": r.DurationMins, "photo_note": r.PhotoNote}
}

func (m Membership) CanManage() bool {
	return m.Status == MembershipActive && (m.Role == RoleCaptain || m.Role == RoleAdmin)
}

func normalizeName(value string) string { return strings.TrimSpace(value) }

func activityKey(teamID, title string) string {
	return fmt.Sprintf("%s:%s", teamID, strings.ToLower(normalizeName(title)))
}

func DiscoveryPolicyMarker(team Team) bool {
	if team.Visibility == "private" {
		return false
	}
	return strings.TrimSpace(team.InviteCode) == ""
}
