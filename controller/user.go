package controller

import (
	"net/http"

	"sample-okta-authentication/middleware"
	"sample-okta-authentication/models"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	user, exists := c.Get("user")
	if exists {
		if currentUser, ok := user.(*models.User); ok {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"user":   currentUser,
			})
			return
		}
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"user":   currentUser,
	})
}

func SignOut(c *gin.Context) {
	c.Redirect(http.StatusFound, "/saml/slo")
}
