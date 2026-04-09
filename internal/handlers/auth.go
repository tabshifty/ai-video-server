package handlers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/utils"
)

// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRegisterRequest true "register payload"
// @Success 200 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /auth/register [post]
func (a *API) RegisterAuth(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || len(req.Password) < 6 {
		bad(c, "username/email required and password length must be >= 6")
		return
	}

	exists, err := a.repo.ExistsUsernameOrEmail(c.Request.Context(), req.Username, req.Email)
	if err != nil {
		response.Error(c, 2, err.Error())
		return
	}
	if exists {
		response.Error(c, 3, "username or email already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 4, "failed to hash password")
		return
	}
	user := models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         "user",
	}
	if err := a.repo.CreateUser(c.Request.Context(), user); err != nil {
		if repository.IsUniqueViolation(err) {
			response.Error(c, 3, "username or email already exists")
			return
		}
		response.Error(c, 5, err.Error())
		return
	}

	tokens, err := a.issueTokens(c, user.ID, user.Role)
	if err != nil {
		response.Error(c, 6, err.Error())
		return
	}

	ok(c, gin.H{
		"user_id":       user.ID,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthLoginRequest true "login payload"
// @Success 200 {object} APIResponse
// @Failure 200 {object} APIResponse
// @Router /auth/login [post]
func (a *API) LoginAuth(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	identity := strings.TrimSpace(req.Username)
	if identity == "" {
		identity = strings.TrimSpace(req.Email)
	}
	if identity == "" || req.Password == "" {
		bad(c, "username/email and password are required")
		return
	}

	user, err := a.repo.GetUserByUsernameOrEmail(c.Request.Context(), identity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 7, "invalid credentials")
			return
		}
		response.Error(c, 8, err.Error())
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, 7, "invalid credentials")
		return
	}

	tokens, err := a.issueTokens(c, user.ID, user.Role)
	if err != nil {
		response.Error(c, 9, err.Error())
		return
	}
	ok(c, gin.H{
		"user_id":       user.ID,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (a *API) RefreshAuth(c *gin.Context) {
	refreshToken, err := extractBearer(c.GetHeader("Authorization"))
	if err != nil {
		response.Error(c, 10, "missing refresh token")
		return
	}

	claims, err := utils.ParseAndValidateToken(a.jwtSecret, refreshToken, utils.TokenTypeRefresh)
	if err != nil {
		response.Error(c, 11, "invalid refresh token")
		return
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(c, 11, "invalid refresh token")
		return
	}
	if err := middleware.ValidateRefreshSession(c, a.redis, claims.ID, uid); err != nil {
		response.Error(c, 12, err.Error())
		return
	}
	if err := middleware.RevokeRefreshSession(c, a.redis, claims.ID); err != nil {
		response.Error(c, 13, err.Error())
		return
	}

	user, err := a.repo.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, 14, err.Error())
		return
	}

	tokens, err := a.issueTokens(c, user.ID, user.Role)
	if err != nil {
		response.Error(c, 15, err.Error())
		return
	}
	ok(c, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (a *API) LogoutAuth(c *gin.Context) {
	accessToken, err := extractBearer(c.GetHeader("Authorization"))
	if err != nil {
		response.Error(c, 16, "missing access token")
		return
	}
	claims, err := utils.ParseAndValidateToken(a.jwtSecret, accessToken, utils.TokenTypeAccess)
	if err != nil {
		response.Error(c, 17, "invalid access token")
		return
	}
	if claims.ExpiresAt == nil {
		response.Error(c, 17, "invalid access token")
		return
	}
	if err := middleware.RevokeAccessToken(c, a.redis, claims.ID, claims.ExpiresAt.Time); err != nil {
		response.Error(c, 18, err.Error())
		return
	}
	ok(c, gin.H{"logged_out": true})
}

func (a *API) issueTokens(c *gin.Context, userID uuid.UUID, role string) (models.AuthTokens, error) {
	accessToken, _, _, err := utils.GenerateToken(a.jwtSecret, userID, role, utils.TokenTypeAccess, a.accessTTL)
	if err != nil {
		return models.AuthTokens{}, err
	}
	refreshToken, refreshJTI, refreshExp, err := utils.GenerateToken(a.jwtSecret, userID, role, utils.TokenTypeRefresh, a.refreshTTL)
	if err != nil {
		return models.AuthTokens{}, err
	}
	if err := middleware.StoreRefreshSession(c, a.redis, refreshJTI, userID, refreshExp); err != nil {
		return models.AuthTokens{}, err
	}
	return models.AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func extractBearer(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid bearer header")
	}
	return strings.TrimSpace(parts[1]), nil
}
