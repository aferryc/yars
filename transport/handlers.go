package transport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aferryc/yars/model"
	"github.com/aferryc/yars/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	reconManagerUC *usecase.ReconManager
	listUC         *usecase.ListUsecase
}

func NewHandler(reconManagerUC *usecase.ReconManager, listUC *usecase.ListUsecase) *Handler {
	return &Handler{
		reconManagerUC: reconManagerUC,
		listUC:         listUC,
	}
}

// IndexPage serves the main upload form page
func (h *Handler) IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "YARS - Reconciliation Portal",
	})
}

func (h *Handler) HandleReconManagerUpload(c *gin.Context) {
	// Logic for handling reconciliation
	uploadURLs, err := h.reconManagerUC.GenerateUploadURLs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, uploadURLs)
}

func (h *Handler) HandleReconManagerInitCompilation(c *gin.Context) {
	var req model.CompilerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Parse dates if provided as strings
	if startDateStr := c.PostForm("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			req.StartDate = startDate
		}
	}

	if endDateStr := c.PostForm("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			req.EndDate = endDate
		}
	}

	err := h.reconManagerUC.InitiateCompilation(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Compilation initiated successfully",
	})
}

func (h *Handler) HandleListReconSummary(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offset parameter",
		})
		return
	}
	summaries, err := h.listUC.ListReconSummaries(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summaries)
}

func (h *Handler) HandleListUnmatchedBank(c *gin.Context) {
	taskID := c.Param("task_id")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offset parameter",
		})
		return
	}
	unmatchedBank, err := h.listUC.ListUnmatchedBankStatements(c.Request.Context(), taskID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, unmatchedBank)
}

func (h *Handler) HandleListUnmatchedTransactions(c *gin.Context) {
	taskID := c.Param("task_id")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offset parameter",
		})
		return
	}
	unmatchedTrx, err := h.listUC.ListUnmatchedTransactions(c.Request.Context(), taskID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, unmatchedTrx)
}
