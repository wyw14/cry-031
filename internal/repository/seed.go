package repository

import (
	"time"

	"github.com/wyw/cry031-volunteer/internal/domain"
)

func DemoState(now time.Time) domain.State {
	now = now.UTC().Truncate(time.Second)
	state := domain.NewState()
	state.Users["u-captain"] = domain.User{ID: "u-captain", DisplayName: "林晓队长", Phone: "13800000001", Role: domain.RoleCaptain, Active: true, CreatedAt: now.Add(-180 * 24 * time.Hour)}
	state.Users["u-volunteer"] = domain.User{ID: "u-volunteer", DisplayName: "周宁", Phone: "13800000002", Role: domain.RoleMember, Active: true, CreatedAt: now.Add(-90 * 24 * time.Hour)}
	state.Users["u-new"] = domain.User{ID: "u-new", DisplayName: "陈叶", Phone: "13800000003", Role: domain.RoleResident, Active: true, CreatedAt: now.Add(-30 * 24 * time.Hour)}
	state.Users["u-admin"] = domain.User{ID: "u-admin", DisplayName: "社区管理员", Phone: "13800000004", Role: domain.RoleAdmin, Active: true, CreatedAt: now.Add(-365 * 24 * time.Hour)}
	state.Communities["community-riverside"] = domain.Community{ID: "community-riverside", Name: "滨河社区", Address: "新桥路 18 号", Description: "滨河居民自治社区", CreatedAt: now.Add(-365 * 24 * time.Hour)}
	state.Communities["community-garden"] = domain.Community{ID: "community-garden", Name: "桂园社区", Address: "桂园路 6 号", Description: "老幼友好社区", CreatedAt: now.Add(-300 * 24 * time.Hour)}
	state.Areas["area-elder"] = domain.ServiceArea{ID: "area-elder", Name: "助老服务"}
	state.Areas["area-environment"] = domain.ServiceArea{ID: "area-environment", Name: "环境维护"}
	state.Teams["team-riverside"] = domain.Team{ID: "team-riverside", CommunityID: "community-riverside", Name: "滨河同行小队", Description: "周末助老与公共空间维护", Status: domain.TeamRecruiting, Visibility: "public", ServiceAreaIDs: []string{"area-elder", "area-environment"}, InviteCode: "RIVER-2026", CaptainID: "u-captain", CreatedAt: now.Add(-120 * 24 * time.Hour)}
	state.Teams["team-garden"] = domain.Team{ID: "team-garden", CommunityID: "community-garden", Name: "桂园守望小队", Description: "探访和社区安全巡查", Status: domain.TeamPaused, Visibility: "community", ServiceAreaIDs: []string{"area-elder"}, InviteCode: "GARDEN-2026", CaptainID: "u-admin", CreatedAt: now.Add(-80 * 24 * time.Hour)}
	state.Memberships["member-captain"] = domain.Membership{ID: "member-captain", TeamID: "team-riverside", UserID: "u-captain", Role: domain.RoleCaptain, Status: domain.MembershipActive, JoinedAt: now.Add(-120 * 24 * time.Hour), Events: []domain.MembershipEvent{}}
	state.Memberships["member-volunteer"] = domain.Membership{ID: "member-volunteer", TeamID: "team-riverside", UserID: "u-volunteer", Role: domain.RoleMember, Status: domain.MembershipActive, JoinedAt: now.Add(-60 * 24 * time.Hour), Events: []domain.MembershipEvent{}}
	state.Activities["activity-visit"] = domain.Activity{ID: "activity-visit", TeamID: "team-riverside", Title: "周末长者探访", Description: "分组探访独居长者并登记需求", Status: domain.ActivityPublished, StartAt: now.Add(48 * time.Hour), EndAt: now.Add(52 * time.Hour), Location: "滨河社区服务站", MaterialNeeds: []string{"血压计", "记录夹"}, CreatedBy: "u-captain", CreatedAt: now.Add(-24 * time.Hour)}
	state.Activities["activity-cleanup"] = domain.Activity{ID: "activity-cleanup", TeamID: "team-riverside", Title: "河岸清洁", Description: "清理步道并检查护栏", Status: domain.ActivityCompleted, StartAt: now.Add(-72 * time.Hour), EndAt: now.Add(-68 * time.Hour), Location: "滨河步道", MaterialNeeds: []string{"手套", "垃圾袋"}, CreatedBy: "u-captain", CreatedAt: now.Add(-96 * time.Hour)}
	state.Slots["slot-visit"] = domain.RoleSlot{ID: "slot-visit", ActivityID: "activity-visit", Name: "探访志愿者", Description: "双人一组入户", Capacity: 2, Status: domain.SlotOpen}
	state.Slots["slot-photo"] = domain.RoleSlot{ID: "slot-photo", ActivityID: "activity-visit", Name: "现场记录", Description: "整理照片说明与物资清单", Capacity: 1, Status: domain.SlotOpen}
	state.Risks["risk-railing"] = domain.RiskItem{ID: "risk-railing", ActivityID: "activity-cleanup", TeamID: "team-riverside", Title: "步道护栏松动", Severity: "high", Description: "靠近三号桥的护栏需要隔离并联系维修", OwnerID: "u-captain", Status: domain.RiskInProgress, DueAt: now.Add(-12 * time.Hour)}
	state.FollowUps["followup-railing"] = domain.FollowUp{ID: "followup-railing", RiskID: "risk-railing", Title: "联系物业设置警示带", OwnerID: "u-captain", DueAt: now.Add(6 * time.Hour), CreatedAt: now.Add(-24 * time.Hour)}
	state.Handoffs["handoff-railing"] = domain.Handoff{ID: "handoff-railing", ActivityID: "activity-cleanup", RiskID: "risk-railing", FromUserID: "u-captain", ToUserID: "u-volunteer", Status: domain.HandoffPending, Note: "维修前每日巡查一次", DueAt: now.Add(-2 * time.Hour)}
	state.Announcements["announcement-1"] = domain.Announcement{ID: "announcement-1", TeamID: "team-riverside", AuthorID: "u-captain", Title: "周末活动集合提醒", Body: "请提前十分钟到服务站领取物资。", CreatedAt: now.Add(-6 * time.Hour)}
	return state
}
