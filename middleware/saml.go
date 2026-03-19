package middleware

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"sample-okta-authentication/models"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

func SamlMiddleware(samlSP *samlsp.Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		if samlSP == nil {
			slog.Error("SAML middleware not initialized")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "SAML middleware is not initialized"})
			return
		}

		authorized := false

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorized = true
			c.Request = r
		})

		samlSP.RequireAccount(next).ServeHTTP(c.Writer, c.Request)
		if !authorized {
			slog.Warn("request not authorized via SAML")
			c.Abort()
			return
		}

		user, err := GetCurrentUser(c)
		if err != nil {
			slog.Error("failed to extract current user", "error", err)
			// clear broken session
			if delErr := samlSP.Session.DeleteSession(c.Writer, c.Request); delErr != nil {
				slog.Error("failed to delete broken SAML session", "error", delErr)
			}

			// trigger SAML login again
			c.Redirect(http.StatusFound, "/user/info")

			c.Abort()
			return
			//c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			//return
		}
		c.Set("user", user)

		c.Next()
	}
}

func Saml(cfg *models.Config) (*samlsp.Middleware, error) {
	if cfg == nil {
		slog.Error("config is nil")
		return nil, fmt.Errorf("config is nil")
	}

	certFile := cfg.SAMLSPCertFile
	keyFile := cfg.SAMLSPKeyFile
	rootURLValue := cfg.AppURL
	idpMetadataURLValue := cfg.SAMLIDPMetadataURL

	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		slog.Error("failed to load service provider key pair", "cert", certFile, "key", keyFile, "error", err)
		return nil, fmt.Errorf("load service provider key pair: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		slog.Error("failed to parse service provider certificate", "error", err)
		return nil, fmt.Errorf("parse service provider certificate: %w", err)
	}

	idpMetadataURL, err := url.Parse(idpMetadataURLValue)
	if err != nil {
		slog.Error("failed to parse IdP metadata URL", "url", idpMetadataURLValue, "error", err)
		return nil, fmt.Errorf("parse IdP metadata URL: %w", err)
	}
	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		slog.Error("failed to fetch IdP metadata", "url", idpMetadataURLValue, "error", err)
		return nil, fmt.Errorf("fetch IdP metadata: %w", err)
	}

	rootURL, err := url.Parse(rootURLValue)
	if err != nil {
		slog.Error("failed to parse application URL", "url", rootURLValue, "error", err)
		return nil, fmt.Errorf("parse application URL: %w", err)
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	if err != nil {
		slog.Error("failed to create SAML service provider", "error", err)
		return nil, fmt.Errorf("create SAML service provider: %w", err)
	}

	slog.Info("SAML service provider initialized successfully", "app_url", rootURLValue)

	return samlSP, nil
}

func GetCurrentUser(c *gin.Context) (*models.User, error) {
	// debug saml attributes
	ctx := c.Request.Context()

	slog.Info("resolved SAML attributes",
		"name_id", samlsp.AttributeFromContext(ctx, "name_id"),
		"email", samlsp.AttributeFromContext(ctx, "email"),
		"firstName", samlsp.AttributeFromContext(ctx, "firstName"),
		"lastName", samlsp.AttributeFromContext(ctx, "lastName"),
		"phone", samlsp.AttributeFromContext(ctx, "phone"),
	)

	email := samlsp.AttributeFromContext(c.Request.Context(), "email")
	if email == "" {
		slog.Warn("user not authenticated")
		return nil, fmt.Errorf("user not authenticated")
	}

	return &models.User{
		FirstName: samlsp.AttributeFromContext(c.Request.Context(), "firstName"),
		LastName:  samlsp.AttributeFromContext(c.Request.Context(), "lastName"),
		Email:     email,
		Phone:     samlsp.AttributeFromContext(c.Request.Context(), "phone"),
	}, nil
}

func SignOut(samlSP *samlsp.Middleware, cfg *models.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if samlSP == nil {
			slog.Error("SAML middleware not initialized for signout")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "SAML middleware is not initialized",
			})
			return
		}

		if err := samlSP.Session.DeleteSession(c.Writer, c.Request); err != nil {
			slog.Error("failed to delete local SAML session", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sign out",
			})
			return
		}

		if cfg != nil && cfg.LogoutURL != "" {
			oktaLogoutURL := cfg.LogoutURL + "/login/signout?fromURI=" + cfg.AppURL
			slog.Info("redirecting to logout URL: " + oktaLogoutURL)
			c.Redirect(http.StatusFound, oktaLogoutURL)
			return
		}

		slog.Info("logout URL not configured, redirecting to root")
		c.Redirect(http.StatusFound, "/")
	}
}
