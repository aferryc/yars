package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/aferryc/yars/transport"
)

// SetupRouter initializes the router and sets up the routes for the application.
func SetupRouter(handler transport.Handler) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("assets/templates/*")

	// Define routes
	router.GET("/", handler.IndexPage)
	router.Static("/static", "./assets/static")
	api := router.Group("/api")
	{
		api.GET("/reconciliation/upload", handler.HandleReconManagerUpload)
		api.POST("/reconciliation", handler.HandleReconManagerInitCompilation)
		api.GET("/reconciliation/summary/list", handler.HandleListReconSummary)
		api.GET("/reconciliation/summary/:task_id/bank", handler.HandleListUnmatchedBank)
		api.GET("/reconciliation/summary/:task_id/transaction", handler.HandleListUnmatchedTransactions)
	}
	return router
}
