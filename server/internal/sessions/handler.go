package sessions

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bizk/system-design-dojo/server/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	imageField = "images"
	audioField = "audio"
)

type Handler struct {
	store       *Store
	storage     *minio.Client
	bucket      string
	maxFileSize int64
}

func NewHandler(store *Store, client *minio.Client, bucket string, maxFileSize int64) *Handler {
	return &Handler{store: store, storage: client, bucket: bucket, maxFileSize: maxFileSize}
}

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/sessions", handler.list)
	router.POST("/sessions", handler.create)
	router.GET("/sessions/:id", handler.get)
	router.PATCH("/sessions/:id", handler.update)
	router.DELETE("/sessions/:id", handler.delete)
	router.GET("/sessions/:id/media/:mediaId", handler.downloadMedia)
	router.DELETE("/sessions/:id/media/:mediaId", handler.deleteMedia)
}

func (h *Handler) list(c *gin.Context) {
	sessions, err := h.store.List()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, responseList(sessions))
}

func (h *Handler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	session, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, h.response(session))
}

func (h *Handler) create(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(h.maxFileSize); err != nil {
		errorResponse(c, http.StatusBadRequest, errors.New("invalid multipart form"))
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		errorResponse(c, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	status := Status(c.PostForm("status"))
	if status == "" {
		status = StatusDraft
	}
	if !validStatus(status) {
		errorResponse(c, http.StatusBadRequest, errors.New("status must be draft, active, or completed"))
		return
	}
	files, err := h.files(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err)
		return
	}

	session := newSession(title, c.PostForm("description"), c.PostForm("content"), status)
	if err := h.store.Create(&session); err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.addFiles(c, session.ID, files); err != nil {
		_ = h.store.Delete(session.ID)
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	session, err = h.store.Get(session.ID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, h.response(session))
}

func (h *Handler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := c.Request.ParseMultipartForm(h.maxFileSize); err != nil {
		errorResponse(c, http.StatusBadRequest, errors.New("invalid multipart form"))
		return
	}
	fields := map[string]any{}
	if _, exists := c.GetPostForm("title"); exists {
		title := strings.TrimSpace(c.PostForm("title"))
		if title == "" {
			errorResponse(c, http.StatusBadRequest, errors.New("title cannot be empty"))
			return
		}
		fields["title"] = title
	}
	for _, field := range []string{"description", "content"} {
		if value, exists := c.GetPostForm(field); exists {
			fields[field] = value
		}
	}
	if value, exists := c.GetPostForm("status"); exists {
		status := Status(value)
		if !validStatus(status) {
			errorResponse(c, http.StatusBadRequest, errors.New("status must be draft, active, or completed"))
			return
		}
		fields["status"] = status
	}
	files, err := h.files(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err)
		return
	}
	if len(fields) == 0 && len(files) == 0 {
		errorResponse(c, http.StatusBadRequest, errors.New("at least one field or media file is required"))
		return
	}
	if len(fields) > 0 {
		fields["updated_at"] = time.Now().UTC()
		if err := h.store.Update(id, fields); err != nil {
			if errors.Is(err, ErrNotFound) {
				errorResponse(c, http.StatusNotFound, err)
				return
			}
			errorResponse(c, http.StatusInternalServerError, err)
			return
		}
	}
	if err := h.addFiles(c, id, files); err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	session, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, h.response(session))
}

func (h *Handler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	session, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.store.Delete(id); err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	for _, media := range session.Media {
		_ = storage.Remove(c.Request.Context(), h.storage, h.bucket, media.ObjectKey)
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) deleteMedia(c *gin.Context) {
	sessionID, ok := parseID(c)
	if !ok {
		return
	}
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, errors.New("invalid media id"))
		return
	}
	media, err := h.store.GetMedia(sessionID, mediaID)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.store.DeleteMedia(sessionID, mediaID); err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	_ = storage.Remove(c.Request.Context(), h.storage, h.bucket, media.ObjectKey)
	c.Status(http.StatusNoContent)
}

func (h *Handler) downloadMedia(c *gin.Context) {
	sessionID, ok := parseID(c)
	if !ok {
		return
	}
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, errors.New("invalid media id"))
		return
	}
	media, err := h.store.GetMedia(sessionID, mediaID)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	object, err := h.storage.GetObject(c.Request.Context(), h.bucket, media.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer object.Close()
	if _, err := object.Stat(); err != nil {
		errorResponse(c, http.StatusNotFound, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(media.Filename)))
	c.DataFromReader(http.StatusOK, media.Size, media.ContentType, object, map[string]string{})
}

type uploadedFile struct {
	header *multipart.FileHeader
	type_  string
}

func (h *Handler) files(c *gin.Context) ([]uploadedFile, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	var files []uploadedFile
	for _, field := range []struct {
		name  string
		type_ string
	}{{imageField, "image"}, {audioField, "audio"}} {
		for _, header := range form.File[field.name] {
			if header.Size <= 0 || header.Size > h.maxFileSize {
				return nil, fmt.Errorf("%s files must be between 1 byte and %d bytes", field.name, h.maxFileSize)
			}
			if !allowedContentType(field.type_, header.Header.Get("Content-Type")) {
				return nil, fmt.Errorf("unsupported %s content type", field.name)
			}
			files = append(files, uploadedFile{header: header, type_: field.type_})
		}
	}
	return files, nil
}

func (h *Handler) addFiles(c *gin.Context, sessionID uuid.UUID, files []uploadedFile) error {
	var uploaded []string
	for _, file := range files {
		body, err := file.header.Open()
		if err != nil {
			return err
		}
		key := fmt.Sprintf("sessions/%s/%s-%s", sessionID, uuid.New(), filepath.Base(file.header.Filename))
		contentType := file.header.Header.Get("Content-Type")
		err = storage.Put(c.Request.Context(), h.storage, h.bucket, key, contentType, file.header.Size, body)
		_ = body.Close()
		if err != nil {
			for _, uploadedKey := range uploaded {
				_ = storage.Remove(c.Request.Context(), h.storage, h.bucket, uploadedKey)
			}
			return err
		}
		media := &Media{ID: uuid.New(), SessionID: sessionID, Type: file.type_, ObjectKey: key, Filename: filepath.Base(file.header.Filename), ContentType: contentType, Size: file.header.Size}
		if err := h.store.AddMedia(media); err != nil {
			_ = storage.Remove(c.Request.Context(), h.storage, h.bucket, key)
			for _, uploadedKey := range uploaded {
				_ = storage.Remove(c.Request.Context(), h.storage, h.bucket, uploadedKey)
			}
			return err
		}
		uploaded = append(uploaded, key)
	}
	return nil
}

func (h *Handler) response(session Session) gin.H {
	return sessionResponse(session, cURL)
}

func responseList(sessions []Session) []gin.H {
	result := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, sessionResponse(session, cURL))
	}
	return result
}

func sessionResponse(session Session, mediaURL func(uuid.UUID, uuid.UUID) string) gin.H {
	images := make([]gin.H, 0)
	audio := make([]gin.H, 0)
	for _, media := range session.Media {
		item := gin.H{"id": media.ID, "type": media.Type, "filename": media.Filename, "contentType": media.ContentType, "size": media.Size, "url": mediaURL(session.ID, media.ID), "transcription": media.Transcription}
		if media.Type == "image" {
			images = append(images, item)
		} else {
			audio = append(audio, item)
		}
	}
	return gin.H{"id": session.ID, "title": session.Title, "description": session.Description, "status": session.Status, "content": session.Content, "images": images, "audio": audio, "createdAt": session.CreatedAt, "updatedAt": session.UpdatedAt}
}

func cURL(sessionID, mediaID uuid.UUID) string {
	return fmt.Sprintf("/api/sessions/%s/media/%s", sessionID, mediaID)
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, errors.New("invalid session id"))
		return uuid.Nil, false
	}
	return id, true
}

func errorResponse(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}

func allowedContentType(mediaType, contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if mediaType == "image" {
		return strings.HasPrefix(contentType, "image/")
	}
	return strings.HasPrefix(contentType, "audio/")
}
