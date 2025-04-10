package router

import (
	"errors"
	"fmt"
	"github.com/csBenClarkson/url-shortener/store"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
)

type registerURLForm struct {
	Url string `json:"url" binding:"required"`
}

type loginForm struct {
	Password string `json:"password" binding:"required"`
}

func getIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (m storageModel) getURL(c *gin.Context) {
	digest := c.Param("digest")
	url, err := m.storage.GetOriginalURL(c.Request.Context(), digest)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.HTML(http.StatusNotFound, "404.html", nil)
		} else if errors.Is(err, store.ErrDBFails) {
			c.HTML(http.StatusServiceUnavailable, "DBFails.html", nil)
			slog.Error("Database fails: %v", err)
		} else {
			c.HTML(http.StatusInternalServerError, "500.html", nil)
			slog.Error(err.Error())
		}
		return
	}
	c.Redirect(http.StatusFound, "https://"+url)
}

func getLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func postLogin(c *gin.Context) {
	var form loginForm
	err := c.ShouldBindJSON(&form)
	if err != nil {
		c.HTML(http.StatusBadRequest, "400.html", nil)
		return
	}

	secret := os.Getenv("SHORTENER_SECRET")
	if secret == form.Password {
		token, err := generateAccessToken(secret)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", nil)
			slog.Error(err.Error())
			return
		}
		c.SetCookie(
			"access_token",
			token,
			int(jwtExpireTime.Seconds()),
			"/",
			"",
			true,
			true,
		)
		c.JSON(http.StatusOK, gin.H{
			"result":   "ok",
			"redirect": "/admin/register",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"result": "err",
			"detail": "Password is not correct.",
		})
	}
}

func getRegisterURL(c *gin.Context) {
	token, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	tokenValid, err := parseAccessToken(token, os.Getenv("SHORTENER_SECRET"))
	if !tokenValid || err != nil {
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	c.HTML(http.StatusOK, "register.html", nil)
}

func (m storageModel) postRegisterURL(c *gin.Context) {
	token, err := c.Cookie("access_token")
	if err != nil {
		c.HTML(http.StatusOK, "login.html", nil)
		return
	}
	tokenValid, err := parseAccessToken(token, os.Getenv("SHORTENER_SECRET"))
	if !tokenValid || err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result": "err",
			"detail": "Not login or token expired.",
		})
		return
	}

	var form registerURLForm
	err = c.ShouldBindJSON(&form)
	if err != nil {
		c.HTML(http.StatusBadRequest, "400.html", nil)
		return
	}
	digest, err := m.storage.StoreURL(c.Request.Context(), form.Url)

	if err != nil {
		if errors.Is(err, store.ErrURLExists) {
			c.JSON(http.StatusOK, gin.H{
				"result": "URL already exists",
				"url":    form.Url,
				"digest": digest,
			})
			slog.Info(fmt.Sprintf("Attempting to register an existing URL: %v", form.Url))
		} else if errors.Is(err, store.ErrDBFails) {
			c.HTML(http.StatusServiceUnavailable, "DBFails.html", nil)
			slog.Error("Database fails: %v", err)
			return
		} else if errors.Is(err, store.ErrTooLucky) {
			c.JSON(http.StatusOK, gin.H{
				"result": "Too lucky",
				"detail": err.Error(),
			})
			slog.Error(err.Error())
		} else {
			c.HTML(http.StatusInternalServerError, "500.html", nil)
			slog.Error(err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
		"digest": digest,
	})
	slog.Info(fmt.Sprintf("Successfully register URL: %v with digest: %v", form.Url, digest))
}

func notFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", nil)
}
