package handlers

import (
	"log/slog"
	"net/http"

	services "personal-bookkeeping/internal/app/service"

	"github.com/gin-gonic/gin"
)

type OCRHandler struct {
	endpoint string
}

func NewOCRHandler(endpoint string) *OCRHandler {
	return &OCRHandler{endpoint: endpoint}
}

// RecognizeReceipt  godoc
// @Summary      拍照识别小票
// @Tags         ocr
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        image formData file true "小票图片 (jpg/png)"
// @Success      200 {object} Response{data=services.OCRResult}
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

	result, err := services.RecognizeReceipt(h.endpoint, file, header.Filename)
	if err != nil {
		slog.Error("ocr failed", "error", err)
		InternalError(c, "识别失败，请稍后重试")
		return
	}

	RespondJSON(c, http.StatusOK, result)
}
