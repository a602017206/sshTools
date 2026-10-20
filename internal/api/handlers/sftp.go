package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"AHaSSHTools/internal/api/dto"
	"AHaSSHTools/internal/service"
	"github.com/gin-gonic/gin"
)

// SFTPHandler exposes per-session SFTP operations over HTTP.
// 桌面端走 Wails 绑定直接调服务层；移动端没有绑定通道，
// 这里把浏览/操作/单文件传输按 REST 暴露，传输走同步接口在请求内完成。
type SFTPHandler struct {
	service *service.SFTPService
}

func NewSFTPHandler(s *service.SFTPService) *SFTPHandler {
	return &SFTPHandler{service: s}
}

func (h *SFTPHandler) ready(c *gin.Context) bool {
	if h.service == nil {
		c.JSON(http.StatusNotImplemented, dto.NewErrorMessageResponse("sftp service not available"))
		return false
	}
	return true
}

func requirePath(c *gin.Context) (string, bool) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("path is required"))
		return "", false
	}
	return path, true
}

// List handles GET /api/v1/sessions/:id/sftp/list?path=
// path 为空时由服务层按会话当前目录（远端家目录）处理。
func (h *SFTPHandler) List(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	files, err := h.service.ListFiles(c.Param("id"), c.Query("path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(files))
}

// Pwd handles GET /api/v1/sessions/:id/sftp/pwd
// 返回会话当前目录，供客户端在列空目录时定位自己所在位置。
func (h *SFTPHandler) Pwd(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	pwd, err := h.service.GetCurrentPath(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{"path": pwd}))
}

// Stat handles GET /api/v1/sessions/:id/sftp/stat?path=
func (h *SFTPHandler) Stat(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	path, ok := requirePath(c)
	if !ok {
		return
	}
	info, err := h.service.GetFileInfo(c.Param("id"), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(info))
}

type mkdirRequest struct {
	Path string `json:"path" binding:"required"`
}

// Mkdir handles POST /api/v1/sessions/:id/sftp/mkdir
func (h *SFTPHandler) Mkdir(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	var req mkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}
	if err := h.service.CreateDirectory(c.Param("id"), req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Directory created"))
}

type renameRequest struct {
	OldPath string `json:"old_path" binding:"required"`
	NewPath string `json:"new_path" binding:"required"`
}

// Rename handles POST /api/v1/sessions/:id/sftp/rename
func (h *SFTPHandler) Rename(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	var req renameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}
	if err := h.service.RenameFile(c.Param("id"), req.OldPath, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Renamed"))
}

// Delete handles DELETE /api/v1/sessions/:id/sftp?path=
func (h *SFTPHandler) Delete(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	path, ok := requirePath(c)
	if !ok {
		return
	}
	if err := h.service.DeleteFile(c.Param("id"), path); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Deleted"))
}

// ReadText handles GET /api/v1/sessions/:id/sftp/read?path=
func (h *SFTPHandler) ReadText(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	path, ok := requirePath(c)
	if !ok {
		return
	}
	file, err := h.service.ReadTextFile(c.Param("id"), path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(file))
}

type writeTextRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content"`
}

// WriteText handles PUT /api/v1/sessions/:id/sftp/write
func (h *SFTPHandler) WriteText(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	var req writeTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse(invalidRequestBodyMessage))
		return
	}
	if err := h.service.WriteTextFile(c.Param("id"), req.Path, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessMessageResponse("Saved"))
}

// Upload handles POST /api/v1/sessions/:id/sftp/upload?dir=<remoteDir>
// 表单字段名 file。先落应用沙盒临时文件，再同步 SFTP 上传，请求内返回结果。
func (h *SFTPHandler) Upload(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	remoteDir := c.Query("dir")
	if remoteDir == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("dir is required"))
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorMessageResponse("file is required"))
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "sshtools-upload-*"+filepath.Ext(fileHeader.Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	tmp.Close()

	// UploadFileSync 以本地文件名为远端名，需要保留用户原始文件名，
	// 因此按原名再落一份到以原名命名的临时路径（Base 防目录穿越）。
	namedPath := filepath.Join(filepath.Dir(tmpPath), filepath.Base(fileHeader.Filename))
	if err := os.Rename(tmpPath, namedPath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	defer os.Remove(namedPath)

	remotePath, err := h.service.UploadFileSync(c.Param("id"), namedPath, remoteDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{"remote_path": remotePath}))
}

// Download handles GET /api/v1/sessions/:id/sftp/download?path=
// 先同步下载到沙盒临时目录，再以附件形式回传给客户端并清理。
func (h *SFTPHandler) Download(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	path, ok := requirePath(c)
	if !ok {
		return
	}

	tmpDir, err := os.MkdirTemp("", "sshtools-download-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	defer os.RemoveAll(tmpDir)

	localPath, err := h.service.DownloadFileSync(c.Param("id"), path, tmpDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
		return
	}
	c.FileAttachment(localPath, filepath.Base(localPath))
}
