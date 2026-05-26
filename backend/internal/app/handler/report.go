package handlers

import (
	"net/http"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/app/repository"
	services "personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// GenerateReport  godoc
// @Summary      生成财务报表 PDF
// @Tags         report
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        period    query string true "周期: monthly | quarterly"
// @Param        date      query string true "日期: YYYY-MM (月度) 或 YYYY-MM (季度)"
// @Success      200 {file} binary
// @Router       /ledgers/{ledger_id}/report [get]
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")
	period := c.Query("period")
	date := c.Query("date")

	if period == "" {
		period = "monthly"
	}
	if date == "" {
		BadRequest(c, "date is required (YYYY-MM)")
		return
	}
	if period != "monthly" && period != "quarterly" {
		BadRequest(c, "period must be monthly or quarterly")
		return
	}

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", lid, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	data, err := services.BuildReportData(database.GetDB(), lid, user.ID, services.ReportPeriod(period), date)
	if err != nil {
		InternalError(c, "failed to build report data: "+err.Error())
		return
	}

	pdfBytes, err := services.GenerateReportPDF(data)
	if err != nil {
		InternalError(c, "failed to generate PDF: "+err.Error())
		return
	}

	filename := ledger.Name + "_" + date + "_report.pdf"
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ReportPreview  godoc
// @Summary      获取报表数据预览（JSON）
// @Tags         report
// @Produce      json
// @Security     BearerAuth
// @Param        ledger_id path string true "账本 ID"
// @Param        period    query string true "周期: monthly | quarterly"
// @Param        date      query string true "日期: YYYY-MM"
// @Success      200 {object} Response{data=services.ReportData}
// @Router       /ledgers/{ledger_id}/report/preview [get]
func (h *ReportHandler) ReportPreview(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")
	period := c.Query("period")
	date := c.Query("date")

	if date == "" {
		BadRequest(c, "date is required (YYYY-MM)")
		return
	}
	if period == "" {
		period = "monthly"
	}

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	var ledger models.Ledger
	if err := database.GetDB().Where("id = ? AND user_id = ?", lid, user.ID).First(&ledger).Error; err != nil {
		NotFound(c, "ledger not found")
		return
	}

	data, err := services.BuildReportData(database.GetDB(), lid, user.ID, services.ReportPeriod(period), date)
	if err != nil {
		InternalError(c, "failed to build report data: "+err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, data)
}
