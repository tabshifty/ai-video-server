package handlers

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"

	"video-server/internal/response"
	"video-server/internal/services"
)

type archiveWorkspaceService interface {
	GetArchiveWorkspace(context.Context, uuid.UUID) (services.ArchiveWorkspaceResponse, error)
	SaveArchiveWorkspace(context.Context, uuid.UUID, services.ArchiveWorkspaceSave) (services.ArchiveWorkspaceResponse, error)
	SubmitArchiveWorkspace(context.Context, uuid.UUID, services.ArchiveWorkspaceSubmit) (services.ArchiveWorkspaceResponse, error)
	GenerateArchiveWorkspaceGroups(context.Context, uuid.UUID, int64) (services.ArchiveWorkspaceResponse, error)
}

func (a *API) archiveWorkspace(c *gin.Context) (archiveWorkspaceService, uuid.UUID, bool) {
	svc, ok := a.archiveImportSvc.(archiveWorkspaceService)
	if !ok {
		response.Error(c, 503, "入库工作台暂不可用")
		return nil, uuid.Nil, false
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		bad(c, "批次标识无效")
		return nil, uuid.Nil, false
	}
	return svc, id, true
}
func writeArchiveWorkspace(c *gin.Context, r services.ArchiveWorkspaceResponse, err error) {
	if err != nil {
		code := 1071
		if errors.Is(err, services.ErrArchiveWorkspaceConflict) {
			code = 409
		}
		response.Error(c, code, err.Error())
		return
	}
	response.JSON(c, 0, "ok", r)
}

// AdminArchiveWorkspace loads an import workspace without changing physical files.
func (a *API) AdminArchiveWorkspace(c *gin.Context) {
	s, id, ok := a.archiveWorkspace(c)
	if !ok {
		return
	}
	r, err := s.GetArchiveWorkspace(c.Request.Context(), id)
	writeArchiveWorkspace(c, r, err)
}

// AdminSaveArchiveWorkspace saves a versioned draft.
func (a *API) AdminSaveArchiveWorkspace(c *gin.Context) {
	s, id, ok := a.archiveWorkspace(c)
	if !ok {
		return
	}
	var in services.ArchiveWorkspaceSave
	if c.ShouldBindJSON(&in) != nil {
		bad(c, "工作区草稿格式无效")
		return
	}
	r, err := s.SaveArchiveWorkspace(c.Request.Context(), id, in)
	writeArchiveWorkspace(c, r, err)
}

// AdminSubmitArchiveWorkspace persists selected work for background processing.
func (a *API) AdminSubmitArchiveWorkspace(c *gin.Context) {
	s, id, ok := a.archiveWorkspace(c)
	if !ok {
		return
	}
	var in services.ArchiveWorkspaceSubmit
	if c.ShouldBindJSON(&in) != nil {
		bad(c, "入库选择格式无效")
		return
	}
	r, err := s.SubmitArchiveWorkspace(c.Request.Context(), id, in)
	writeArchiveWorkspace(c, r, err)
}

// AdminGenerateArchiveWorkspaceGroups adds missing directory-based groups.
func (a *API) AdminGenerateArchiveWorkspaceGroups(c *gin.Context) {
	s, id, ok := a.archiveWorkspace(c)
	if !ok {
		return
	}
	var in struct {
		Version int64 `json:"version"`
	}
	if c.ShouldBindJSON(&in) != nil {
		bad(c, "草稿版本无效")
		return
	}
	r, err := s.GenerateArchiveWorkspaceGroups(c.Request.Context(), id, in.Version)
	writeArchiveWorkspace(c, r, err)
}

func (a *API) legacyArchiveWorkspaceGuard(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc, ok := a.archiveImportSvc.(interface {
			AcquireLegacyWorkspaceGuard(context.Context, string, uuid.UUID) (func(), error)
		})
		if !ok {
			c.Next()
			return
		}
		rawID := c.Param("id")
		if kind == "targets" {
			var body struct {
				Targets []struct {
					ID string `json:"id"`
				} `json:"targets"`
			}
			if c.ShouldBindBodyWith(&body, binding.JSON) != nil || len(body.Targets) == 0 {
				bad(c, "请选择文件")
				c.Abort()
				return
			}
			rawID = body.Targets[0].ID
		}
		id, err := uuid.Parse(rawID)
		if err != nil {
			bad(c, "标识无效")
			c.Abort()
			return
		}
		release, err := svc.AcquireLegacyWorkspaceGuard(c.Request.Context(), mapLegacyKind(kind), id)
		if err != nil {
			response.Error(c, 409, err.Error())
			c.Abort()
			return
		}
		defer release()
		c.Next()
	}
}

func mapLegacyKind(kind string) string {
	if kind == "targets" {
		return "file"
	}
	return kind
}
