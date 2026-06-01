package handler

import (
	"net/http"

	"personal-bookkeeping/internal/app/models"
	service "personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
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
		BadRequest(c, "date is required (YYYY-MM for monthly/quarterly, YYYY for yearly)")
		return
	}
	if period != "monthly" && period != "quarterly" && period != "yearly" {
		BadRequest(c, "period must be monthly, quarterly, or yearly")
		return
	}

	lid, err := uuid.Parse(ledgerID)
	if err != nil {
		BadRequest(c, "invalid ledger_id")
		return
	}

	data, err := h.svc.BuildReportData(lid, user.ID, service.ReportPeriod(period), date)
	if err != nil {
		InternalError(c, "failed to build report data: "+err.Error())
		return
	}

	pdfBytes, err := service.GenerateReportPDF(data)
	if err != nil {
		InternalError(c, "failed to generate PDF: "+err.Error())
		return
	}

	filename := data.LedgerName + "_" + date + "_report.pdf"
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
// @Success      200 {object} Response{data=service.ReportData}
// @Router       /ledgers/{ledger_id}/report/preview [get]
func (h *ReportHandler) ReportPreview(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ledgerID := c.Param("ledger_id")
	period := c.Query("period")
	date := c.Query("date")

	if date == "" {
		BadRequest(c, "date is required (YYYY-MM for monthly/quarterly, YYYY for yearly)")
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

	data, err := h.svc.BuildReportData(lid, user.ID, service.ReportPeriod(period), date)
	if err != nil {
		InternalError(c, "failed to build report data: "+err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, data)
}
