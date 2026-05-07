package membership

import (
	"fmt"

	"github.com/1204244136/MDA/agent/go-service/pkg/i18n"
	"github.com/1204244136/MDA/agent/go-service/pkg/maafocus"
	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// MembershipChecker checks if the user has membership before executing member-only tasks.
type MembershipChecker struct{}

// OnTaskerTask handles tasker task events.
func (c *MembershipChecker) OnTaskerTask(tasker *maa.Tasker, event maa.EventStatus, detail maa.TaskerTaskDetail) {
	if event != maa.EventStatusStarting {
		return
	}

	if detail.Entry == "MaaTaskerPostStop" {
		log.Debug().Msg("Received PostStop event, skipping membership check")
		return
	}

	// Ensure member-only entries are loaded from task files
	LoadMemberOnlyEntries()

	// Check if this task requires membership
	if !IsMemberOnly(detail.Entry) {
		log.Debug().Str("entry", detail.Entry).Msg("Task is free, skipping membership check")
		return
	}

	log.Info().
		Uint64("task_id", detail.TaskID).
		Str("entry", detail.Entry).
		Msg("Member-only task detected, checking membership status")

	status := GetMembershipStatus()

	// Log unsupported tier warning
	if status.UnsupportedTier {
		log.Warn().
			Str("tier", status.MembershipType).
			Msg("铜/银会员等级在 MDA 中不再受支持，请升级至金Doro会员或以上")
	}

	// Check if user has sufficient membership level
	if status.IsMember {
		log.Info().
			Str("tier", status.MembershipType).
			Int("level", status.UserLevel).
			Str("expiry", status.VirtualExpiry).
			Msg("Membership verified, task allowed")
		return
	}

	// Non-member trying to run a member-only task
	sponsorURL := fmt.Sprintf(
		"https://doropay.top?cpu=%s&uuid=%s&bios=%s&board=%s&disk=%s&guid=%s",
		status.DeviceCode.CPUHash,
		status.DeviceCode.UUIDHash,
		status.DeviceCode.BIOSHash,
		status.DeviceCode.BoardHash,
		status.DeviceCode.DiskHash,
		status.DeviceCode.GUIDHash,
	)
	log.Warn().
		Str("entry", detail.Entry).
		Str("tier", status.MembershipType).
		Int("level", status.UserLevel).
		Str("sponsor_url", sponsorURL).
		Msg(i18n.T("tasker.membership_warning.non_member_log"))

	c.stopWithWarning(tasker, status, detail.Entry)
}

func (c *MembershipChecker) stopWithWarning(tasker *maa.Tasker, status *MembershipStatus, entry string) {
	maafocus.PrintLargeContentTrimNewline(
		i18n.RenderHTML("tasker.membership_warning", buildWarningData(status, entry)),
	)
	tasker.PostStop()
}

func buildWarningData(status *MembershipStatus, entry string) map[string]any {
	tierDisplay := status.MembershipType
	if tierDisplay == "" || tierDisplay == "普通用户" {
		tierDisplay = i18n.T("tasker.membership_warning.tier_free")
	}

	return map[string]any{
		"CurrentTier": tierDisplay,
		"TaskEntry":   entry,
		"MinLevel":    fmt.Sprintf("%d", minMemberLevel),
	}
}
