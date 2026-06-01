package service

import (
	"errors"
	"time"

	"personal-bookkeeping/internal/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------- sentinel errors ----------

var (
	ErrMemberNotFound    = errors.New("member not found")
	ErrMemberExists      = errors.New("user is already a member of this ledger")
	ErrCannotRemoveOwner = errors.New("cannot remove the owner of the ledger")
	ErrCannotDemoteOwner = errors.New("cannot change the role of the owner")
	ErrInviteExpired     = errors.New("invite token expired or invalid")
	ErrNotInvited        = errors.New("no pending invitation found")
	ErrSelfRemove        = errors.New("owner cannot leave the ledger; transfer ownership first")
)

// ---------- request / response types ----------

type MemberInfo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ---------- service ----------

type MemberService struct {
	db *gorm.DB
}

func NewMemberService(s *Service) *MemberService {
	return &MemberService{db: s.DB}
}

// ListMembers returns all members of a ledger.
func (s *MemberService) ListMembers(ledgerID uuid.UUID) ([]MemberInfo, error) {
	var members []models.LedgerMember
	if err := s.db.Preload("User").
		Where("ledger_id = ?", ledgerID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}

	info := make([]MemberInfo, len(members))
	for i, m := range members {
		info[i] = MemberInfo{
			ID:        m.ID,
			UserID:    m.UserID,
			Username:  m.User.Username,
			Role:      m.Role,
			JoinedAt:  m.JoinedAt,
			CreatedAt: m.CreatedAt,
		}
	}
	return info, nil
}

// GetMemberRole returns the role of a user in a ledger.
func (s *MemberService) GetMemberRole(ledgerID, userID uuid.UUID) (string, error) {
	var member models.LedgerMember
	if err := s.db.Where("ledger_id = ? AND user_id = ?", ledgerID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrMemberNotFound
		}
		return "", err
	}
	return member.Role, nil
}

// InviteMember adds a user to a ledger. Only owner/admin can invite.
func (s *MemberService) InviteMember(ledgerID, invitedBy uuid.UUID, username string) (*MemberInfo, error) {
	// Find the target user
	var targetUser models.User
	if err := s.db.Where("username = ?", username).First(&targetUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check not already a member
	var existing int64
	s.db.Model(&models.LedgerMember{}).
		Where("ledger_id = ? AND user_id = ?", ledgerID, targetUser.ID).
		Count(&existing)
	if existing > 0 {
		return nil, ErrMemberExists
	}

	member := models.LedgerMember{
		LedgerID:  ledgerID,
		UserID:    targetUser.ID,
		Role:      models.RoleMember,
		InvitedBy: &invitedBy,
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(&member).Error; err != nil {
		return nil, err
	}

	// Reload with User preloaded
	s.db.Preload("User").First(&member, member.ID)

	return &MemberInfo{
		ID:        member.ID,
		UserID:    member.UserID,
		Username:  member.User.Username,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
		CreatedAt: member.CreatedAt,
	}, nil
}

// RemoveMember removes a user from a ledger. Cannot remove the owner.
func (s *MemberService) RemoveMember(ledgerID uuid.UUID, targetUserID uuid.UUID) error {
	var member models.LedgerMember
	if err := s.db.Where("ledger_id = ? AND user_id = ?", ledgerID, targetUserID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	if member.Role == models.RoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.db.Delete(&member).Error
}

// LeaveLedger removes the current user from a ledger. Owner cannot leave.
func (s *MemberService) LeaveLedger(ledgerID, userID uuid.UUID) error {
	var member models.LedgerMember
	if err := s.db.Where("ledger_id = ? AND user_id = ?", ledgerID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	if member.Role == models.RoleOwner {
		return ErrSelfRemove
	}

	return s.db.Delete(&member).Error
}

// UpdateMemberRole changes a member's role. Only owner can promote/demote.
func (s *MemberService) UpdateMemberRole(ledgerID, targetUserID uuid.UUID, newRole string) error {
	if newRole != models.RoleAdmin && newRole != models.RoleMember {
		return errors.New("invalid role: must be admin or member")
	}

	var member models.LedgerMember
	if err := s.db.Where("ledger_id = ? AND user_id = ?", ledgerID, targetUserID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	if member.Role == models.RoleOwner {
		return ErrCannotDemoteOwner
	}

	return s.db.Model(&member).Update("role", newRole).Error
}

// MigrateLedgerOwnership ensures all existing ledgers have an owner member record.
func (s *MemberService) MigrateLedgerOwnership() error {
	var ledgers []models.Ledger
	if err := s.db.Find(&ledgers).Error; err != nil {
		return err
	}

	for _, l := range ledgers {
		var count int64
		s.db.Model(&models.LedgerMember{}).
			Where("ledger_id = ? AND user_id = ?", l.ID, l.UserID).
			Count(&count)
		if count == 0 {
			member := models.LedgerMember{
				LedgerID: l.ID,
				UserID:   l.UserID,
				Role:     models.RoleOwner,
				JoinedAt: l.CreatedAt,
			}
			if err := s.db.Create(&member).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
