package followup

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/call-notes-ai-service/internal/modules/followup/entities"
	"github.com/google/uuid"
)

type ICore interface {
	DetectFollowUps(ctx context.Context, sessionID uuid.UUID, text string) ([]entities.FollowUpDetection, error)
	GetFollowUps(ctx context.Context, sessionID uuid.UUID) ([]entities.FollowUp, error)
	ConfirmFollowUp(ctx context.Context, req *entities.ConfirmFollowUpRequest) (*entities.FollowUp, error)
}

type Core struct {
	repo     IRepository
	patterns []followUpPattern
}

type followUpPattern struct {
	regex        *regexp.Regexp
	followUpType entities.FollowUpType
	description  string
	defaultDays  int
}

var _ ICore = (*Core)(nil)

func NewCore(repo IRepository) ICore {
	return &Core{
		repo:     repo,
		patterns: buildPatterns(),
	}
}

func (c *Core) DetectFollowUps(ctx context.Context, sessionID uuid.UUID, text string) ([]entities.FollowUpDetection, error) {
	lower := strings.ToLower(text)
	var detections []entities.FollowUpDetection

	for _, p := range c.patterns {
		if p.regex.MatchString(lower) {
			durationDays := p.defaultDays
			if extracted := extractDuration(lower); extracted > 0 {
				durationDays = extracted
			}

			dueDate := time.Now().UTC().AddDate(0, 0, durationDays)

			detection := entities.FollowUpDetection{
				Type:         p.followUpType,
				Description:  p.description,
				RawText:      text,
				DueDate:      &dueDate,
				DurationDays: durationDays,
			}
			detections = append(detections, detection)

			now := time.Now().UTC()
			fu := &entities.FollowUp{
				ID:           uuid.New(),
				SessionID:    sessionID,
				FollowUpType: p.followUpType,
				Description:  p.description,
				RawText:      text,
				DueDate:      &dueDate,
				Status:       entities.StatusDetected,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := c.repo.CreateFollowUp(ctx, fu); err != nil {
				logger.Ctx(ctx).Errorw(constants.LogMsgFollowupCreateFailed,
					constants.LogKeyError, err,
					constants.LogFieldSessionID, sessionID.String(),
				)
			}

			logger.Ctx(ctx).Infow(constants.LogMsgFollowupDetected,
				constants.LogFieldSessionID, sessionID.String(),
				constants.LogFieldFollowupType, string(p.followUpType),
				constants.LogFieldDueDate, dueDate.Format("2006-01-02"),
			)
		}
	}

	return detections, nil
}

func (c *Core) GetFollowUps(ctx context.Context, sessionID uuid.UUID) ([]entities.FollowUp, error) {
	return c.repo.GetFollowUps(ctx, sessionID)
}

func (c *Core) ConfirmFollowUp(ctx context.Context, req *entities.ConfirmFollowUpRequest) (*entities.FollowUp, error) {
	followUpID, err := uuid.Parse(req.FollowUpID)
	if err != nil {
		return nil, err
	}

	newStatus := entities.StatusConfirmed
	if !req.Confirmed {
		newStatus = entities.StatusDismissed
	}

	if err := c.repo.UpdateFollowUpStatus(ctx, followUpID, newStatus, &req.AgentID); err != nil {
		return nil, err
	}

	fu, err := c.repo.GetFollowUp(ctx, followUpID)
	if err != nil {
		return nil, err
	}

	logger.Ctx(ctx).Infow(constants.LogMsgFollowupConfirmed,
		constants.LogFieldFollowupID, followUpID.String(),
		constants.LogFieldFollowupStatus, string(newStatus),
	)

	return fu, nil
}

func buildPatterns() []followUpPattern {
	return []followUpPattern{
		{
			regex:        regexp.MustCompile(`(?i)(come back|follow up|check again|phir aana|dobara aana|check karana|wapas aana)`),
			followUpType: entities.FollowUpCallback,
			description:  "Patient callback for follow-up",
			defaultDays:  entities.DefaultCallbackDays,
		},
		{
			regex:        regexp.MustCompile(`(?i)(blood test|lab test|get tested|khoon ki jaanch|test karwana|x-ray|mri|scan)`),
			followUpType: entities.FollowUpLabTest,
			description:  "Lab test or diagnostic imaging",
			defaultDays:  entities.DefaultLabTestDays,
		},
		{
			regex:        regexp.MustCompile(`(?i)(take this medicine|prescription|refill|dawaai lena|dawai|medicine continue)`),
			followUpType: entities.FollowUpPrescriptionRefill,
			description:  "Prescription or medication refill",
			defaultDays:  entities.DefaultPrescriptionDays,
		},
		{
			regex:        regexp.MustCompile(`(?i)(appointment|schedule|book|specialist|doctor se milo|time fix)`),
			followUpType: entities.FollowUpAppointment,
			description:  "Specialist appointment",
			defaultDays:  entities.DefaultAppointmentDays,
		},
		{
			regex:        regexp.MustCompile(`(?i)(if.*not better|if.*persist|agar theek na ho|kam na ho|if.*worse)`),
			followUpType: entities.FollowUpConditional,
			description:  "Conditional follow-up if symptoms persist",
			defaultDays:  entities.DefaultConditionalDays,
		},
	}
}

var durationRegex = regexp.MustCompile(`(?i)(\d+)\s*(day|days|din|week|weeks|hafte|month|months|mahine)`)

func extractDuration(text string) int {
	matches := durationRegex.FindStringSubmatch(text)
	if len(matches) < 3 {
		return 0
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	unit := strings.ToLower(matches[2])
	switch unit {
	case "day", "days", "din":
		return num
	case "week", "weeks", "hafte":
		return num * 7
	case "month", "months", "mahine":
		return num * 30
	default:
		return 0
	}
}
