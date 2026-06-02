package handler

import (
	"errors"
	"fmt"
	"net/http"

	"personal-bookkeeping/internal/app/models"
	"personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MemberHandler struct {
	memberSvc *service.MemberService
	ledgerSvc *service.LedgerService
}

func NewMemberHandler(memberSvc *service.MemberService, ledgerSvc *service.LedgerService) *MemberHandler {
	return &MemberHandler{memberSvc: memberSvc, ledgerSvc: ledgerSvc}
}

type InviteInput struct {
	Username string `json:"username" binding:"required,max=50"`
}

type UpdateRoleInput struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// ListMembers  godoc
// @Summary      账本成员列表
// @Tags         members
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/members [get]
func (h *MemberHandler) ListMembers(c *gin.Context) {
	ledgerID, err := parseUUID(c.Param("ledger_id"))
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	members, err := h.memberSvc.ListMembers(*ledgerID)
	if err != nil {
		InternalError(c, "failed to list members")
		return
	}

	RespondJSON(c, http.StatusOK, members)
}

// Invite  godoc
// @Summary      邀请用户加入账本
// @Tags         members
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        input body InviteInput true "用户名"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/members [post]
func (h *MemberHandler) Invite(c *gin.Context) {
	ledgerID, err := parseUUID(c.Param("ledger_id"))
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	user := c.MustGet("user").(*models.User)
	role, ok := c.Get("ledger_role")
	if !ok || (role != models.RoleOwner && role != models.RoleAdmin) {
		RespondError(c, http.StatusForbidden, "only owner or admin can invite members")
		return
	}

	var input InviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	member, err := h.memberSvc.InviteMember(*ledgerID, user.ID, input.Username)
	if err != nil {
		if errors.Is(err, service.ErrMemberExists) {
			RespondError(c, http.StatusConflict, "user is already a member")
			return
		}
		if err.Error() == "user not found" {
			NotFound(c, "user not found")
			return
		}
		InternalError(c, "failed to invite member")
		return
	}

	RespondJSON(c, http.StatusOK, member)
}

// Remove  godoc
// @Summary      移除账本成员
// @Tags         members
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        user_id path string true "用户 ID"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/members/{user_id} [delete]
func (h *MemberHandler) Remove(c *gin.Context) {
	ledgerID, err := parseUUID(c.Param("ledger_id"))
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		BadRequest(c, "invalid user_id")
		return
	}

	role, ok := c.Get("ledger_role")
	if !ok || (role != models.RoleOwner && role != models.RoleAdmin) {
		RespondError(c, http.StatusForbidden, "only owner or admin can remove members")
		return
	}

	if err := h.memberSvc.RemoveMember(*ledgerID, *userID); err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			NotFound(c, "member not found")
			return
		}
		if errors.Is(err, service.ErrCannotRemoveOwner) {
			RespondError(c, http.StatusForbidden, "cannot remove the owner")
			return
		}
		InternalError(c, "failed to remove member")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

// Leave  godoc
// @Summary      退出账本
// @Tags         members
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/leave [post]
func (h *MemberHandler) Leave(c *gin.Context) {
	ledgerID, err := parseUUID(c.Param("ledger_id"))
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	user := c.MustGet("user").(*models.User)

	if err := h.memberSvc.LeaveLedger(*ledgerID, user.ID); err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			NotFound(c, "not a member")
			return
		}
		if errors.Is(err, service.ErrSelfRemove) {
			RespondError(c, http.StatusForbidden, "owner cannot leave; transfer ownership first")
			return
		}
		InternalError(c, "failed to leave ledger")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

// UpdateRole  godoc
// @Summary      修改成员角色
// @Tags         members
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        user_id path string true "用户 ID"
// @Param        input body UpdateRoleInput true "新角色"
// @Success      200 {object} Response
// @Router       /ledgers/{ledger_id}/members/{user_id} [put]
func (h *MemberHandler) UpdateRole(c *gin.Context) {
	ledgerID, err := parseUUID(c.Param("ledger_id"))
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}
	targetUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		BadRequest(c, "invalid user_id")
		return
	}

	role, ok := c.Get("ledger_role")
	if !ok || role != models.RoleOwner {
		RespondError(c, http.StatusForbidden, "only owner can change roles")
		return
	}

	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.memberSvc.UpdateMemberRole(*ledgerID, *targetUserID, input.Role); err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			NotFound(c, "member not found")
			return
		}
		if errors.Is(err, service.ErrCannotDemoteOwner) {
			RespondError(c, http.StatusForbidden, "cannot change owner's role")
			return
		}
		InternalError(c, "failed to update role")
		return
	}

	RespondJSON(c, http.StatusOK, nil)
}

func parseUUID(s string) (*uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return &id, nil
}
