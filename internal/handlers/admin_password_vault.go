package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/response"
)

func normalizePasswordVaultRequest(in models.AdminPasswordVaultInput) models.AdminPasswordVaultInput {
	in.Name = strings.Join(strings.Fields(strings.TrimSpace(in.Name)), " ")
	in.Account = strings.TrimSpace(in.Account)
	in.URL = strings.TrimSpace(in.URL)
	in.Note = strings.TrimSpace(in.Note)
	return in
}

func (a *API) AdminPasswordVaultEntries(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)

	items, total, err := a.repo.ListPasswordVaultEntries(c.Request.Context(), c.Query("q"), page, pageSize)
	if err != nil {
		response.Error(c, 1080, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminCreatePasswordVaultEntry(c *gin.Context) {
	if a.passwordVaultCipher == nil {
		response.Error(c, 1081, "password vault cipher unavailable")
		return
	}
	var req models.AdminPasswordVaultInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req = normalizePasswordVaultRequest(req)
	if req.Name == "" {
		bad(c, "名称不能为空")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		bad(c, "密码不能为空")
		return
	}
	ciphertext, err := a.passwordVaultCipher.Encrypt(req.Password)
	if err != nil {
		response.Error(c, 1082, err.Error())
		return
	}
	item, err := a.repo.CreatePasswordVaultEntry(c.Request.Context(), req, ciphertext)
	if err != nil {
		response.Error(c, 1083, err.Error())
		return
	}
	ok(c, item)
}

func (a *API) AdminUpdatePasswordVaultEntry(c *gin.Context) {
	if a.passwordVaultCipher == nil {
		response.Error(c, 1081, "password vault cipher unavailable")
		return
	}
	entryID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid password entry id")
		return
	}
	var req models.AdminPasswordVaultInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req = normalizePasswordVaultRequest(req)
	if req.Name == "" {
		bad(c, "名称不能为空")
		return
	}
	var ciphertext *string
	if strings.TrimSpace(req.Password) != "" {
		encrypted, err := a.passwordVaultCipher.Encrypt(req.Password)
		if err != nil {
			response.Error(c, 1082, err.Error())
			return
		}
		ciphertext = &encrypted
	}
	item, err := a.repo.UpdatePasswordVaultEntry(c.Request.Context(), entryID, req, ciphertext)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "password entry not found")
			return
		}
		response.Error(c, 1084, err.Error())
		return
	}
	ok(c, item)
}

func (a *API) AdminDeletePasswordVaultEntry(c *gin.Context) {
	entryID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid password entry id")
		return
	}
	if err := a.repo.DeletePasswordVaultEntry(c.Request.Context(), entryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "password entry not found")
			return
		}
		response.Error(c, 1085, err.Error())
		return
	}
	ok(c, gin.H{"deleted": true, "id": entryID})
}

func (a *API) AdminPasswordVaultPassword(c *gin.Context) {
	if a.passwordVaultCipher == nil {
		response.Error(c, 1081, "password vault cipher unavailable")
		return
	}
	entryID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid password entry id")
		return
	}
	ciphertext, err := a.repo.GetPasswordVaultCiphertext(c.Request.Context(), entryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "password entry not found")
			return
		}
		response.Error(c, 1086, err.Error())
		return
	}
	password, err := a.passwordVaultCipher.Decrypt(ciphertext)
	if err != nil {
		response.Error(c, 1087, err.Error())
		return
	}
	ok(c, gin.H{"password": password})
}
