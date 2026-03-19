package routes

import (
	"net/http"
	"sample-okta-authentication/controller"
	"sample-okta-authentication/middleware"
	"sample-okta-authentication/models"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine, samlSP *samlsp.Middleware, cfg *models.Config) {
	// Define a simple GET endpoint
	router.GET("/", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "welcome to okta-authentication app",
		})
	})

	// SAML routes
	router.GET("/saml/metadata", gin.WrapH(samlSP))
	router.POST("/saml/acs", gin.WrapH(samlSP))
	router.GET("/saml/sso", gin.WrapH(samlSP))
	router.GET("/signout", middleware.SignOut(samlSP, cfg))

	protected := router.Group("/user")
	protected.Use(middleware.SamlMiddleware(samlSP))
	{
		protected.GET("/info", controller.GetUserInfo)
	}
}
