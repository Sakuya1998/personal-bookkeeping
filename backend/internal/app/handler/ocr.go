package handler

import (
	"log/slog"
	"net/http"

	service "personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
)

type OCRHandler struct {
	endpoint string
	svc      *service.OCRService
}

func NewOCRHandler(endpoint string, svc *service.OCRService) *OCRHandler {
	return &OCRHandler{endpoint: endpoint, svc: svc}
}

// RecognizeReceipt  godoc
// @Summary      拍照识别小票
// @Tags         ocr
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        image formData file true "小票图片 (jpg/png)"
// @Success      200 {object} Response{data=service.OCRResult}
// @Router       /ocr/receipt [post]
func (h *OCRHandler) RecognizeReceipt(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		BadRequest(c, "请上传图片文件")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		BadRequest(c, "仅支持 JPG/PNG 格式")
		return
	}

	// Max 10MB
	if header.Size > 10*1024*1024 {
		BadRequest(c, "图片大小不能超过 10MB")
		return
	}

	result, err := service.RecognizeReceipt(h.endpoint, file, header.Filename)
	if err != nil {
		slog.Error("ocr failed", "error", err)
		InternalError(c, "识别失败，请稍后重试")
		return
	}

	RespondJSON(c, http.StatusOK, result)
}
