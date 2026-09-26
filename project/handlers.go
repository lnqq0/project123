package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{Store: store}
}

// helpers

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func validStatus(status string) bool {
	return status == "new" ||
		status == "in_progress" ||
		status == "done"
}

// writeStoreError выбирает код ответа по ошибке из store.
// Текст ошибки базы клиенту не уходит, только в лог.
func writeStoreError(w http.ResponseWriter, err error, entity string) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, entity+" not found")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusUnprocessableEntity, entity+" already exists")
	case errors.Is(err, ErrInvalidRef):
		writeError(w, http.StatusUnprocessableEntity, "related record not found")
	default:
		log.Printf("%s: %v", entity, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// users
type userCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userPatchRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.GetUsers(r.Context())
	if err != nil {
		writeStoreError(w, err, "users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.Store.GetUser(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req userCreateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusUnprocessableEntity, "name and email are required")
		return
	}

	user, err := h.Store.CreateUser(r.Context(), User{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeStoreError(w, err, "user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req userPatchRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := h.Store.GetUser(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "user")
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			writeError(w, http.StatusUnprocessableEntity, "name cannot be empty")
			return
		}

		user.Name = name
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)

		if email == "" {
			writeError(w, http.StatusUnprocessableEntity, "email cannot be empty")
			return
		}

		user.Email = email
	}

	user, err = h.Store.UpdateUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, err, "user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Store.DeleteUser(r.Context(), id)
	if errors.Is(err, ErrInvalidRef) {
		writeError(w, http.StatusConflict, "user has tickets or comments")
		return
	}

	if err != nil {
		writeStoreError(w, err, "user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// categories

type categoryCreateRequest struct {
	Name string `json:"name"`
}

type categoryPatchRequest struct {
	Name *string `json:"name"`
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Store.GetCategories(r.Context())
	if err != nil {
		writeStoreError(w, err, "categories")
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	category, err := h.Store.GetCategory(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "category")
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryCreateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	category, err := h.Store.CreateCategory(r.Context(), Category{
		Name: req.Name,
	})

	if err != nil {
		writeStoreError(w, err, "category")
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req categoryPatchRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	category, err := h.Store.GetCategory(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "category")
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			writeError(w, http.StatusUnprocessableEntity, "name cannot be empty")
			return
		}

		category.Name = name
	}

	category, err = h.Store.UpdateCategory(r.Context(), category)

	if err != nil {
		writeStoreError(w, err, "category")
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Store.DeleteCategory(r.Context(), id)
	if errors.Is(err, ErrInvalidRef) {
		writeError(w, http.StatusConflict, "category has tickets")
		return
	}

	if err != nil {
		writeStoreError(w, err, "category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// tickets

type ticketCreateRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	UserID      int64   `json:"user_id"`
	CategoryID  int64   `json:"category_id"`
}

type ticketPatchRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	UserID      *int64  `json:"user_id"`
	CategoryID  *int64  `json:"category_id"`
}

func (h *Handler) GetTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.Store.GetTickets(r.Context())

	if err != nil {
		writeStoreError(w, err, "tickets")
		return
	}

	writeJSON(w, http.StatusOK, tickets)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ticket, err := h.Store.GetTicket(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "ticket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req ticketCreateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)

	if req.Title == "" || req.UserID <= 0 || req.CategoryID <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "invalid ticket data")
		return
	}

	if req.Status == "" {
		req.Status = "new"
	}

	if !validStatus(req.Status) {
		writeError(w, http.StatusUnprocessableEntity, "invalid status")
		return
	}

	ticket, err := h.Store.CreateTicket(r.Context(), Ticket{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
	})

	if err != nil {
		writeStoreError(w, err, "ticket")
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req ticketPatchRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ticket, err := h.Store.GetTicket(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "ticket")
		return
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)

		if title == "" {
			writeError(w, http.StatusUnprocessableEntity, "title cannot be empty")
			return
		}

		ticket.Title = title
	}

	if req.Description != nil {
		ticket.Description = req.Description
	}

	if req.Status != nil {
		if !validStatus(*req.Status) {
			writeError(w, http.StatusUnprocessableEntity, "invalid status")
			return
		}

		ticket.Status = *req.Status
	}

	if req.UserID != nil {
		if *req.UserID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "invalid user_id")
			return
		}

		ticket.UserID = *req.UserID
	}

	if req.CategoryID != nil {
		if *req.CategoryID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "invalid category_id")
			return
		}

		ticket.CategoryID = *req.CategoryID
	}

	ticket, err = h.Store.UpdateTicket(r.Context(), ticket)

	if err != nil {
		writeStoreError(w, err, "ticket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (h *Handler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Store.DeleteTicket(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "ticket")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// comments

type commentCreateRequest struct {
	TicketID int64  `json:"ticket_id"`
	UserID   int64  `json:"user_id"`
	Text     string `json:"text"`
}

type commentPatchRequest struct {
	TicketID *int64  `json:"ticket_id"`
	UserID   *int64  `json:"user_id"`
	Text     *string `json:"text"`
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	comments, err := h.Store.GetComments(r.Context())

	if err != nil {
		writeStoreError(w, err, "comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	comment, err := h.Store.GetComment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "comment")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req commentCreateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Text = strings.TrimSpace(req.Text)

	if req.TicketID <= 0 || req.UserID <= 0 || req.Text == "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid comment data")
		return
	}

	comment, err := h.Store.CreateComment(r.Context(), Comment{
		TicketID: req.TicketID,
		UserID:   req.UserID,
		Text:     req.Text,
	})

	if err != nil {
		writeStoreError(w, err, "comment")
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req commentPatchRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	comment, err := h.Store.GetComment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "comment")
		return
	}

	if req.TicketID != nil {
		if *req.TicketID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "invalid ticket_id")
			return
		}

		comment.TicketID = *req.TicketID
	}

	if req.UserID != nil {
		if *req.UserID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "invalid user_id")
			return
		}

		comment.UserID = *req.UserID
	}

	if req.Text != nil {
		text := strings.TrimSpace(*req.Text)

		if text == "" {
			writeError(w, http.StatusUnprocessableEntity, "text cannot be empty")
			return
		}

		comment.Text = text
	}

	comment, err = h.Store.UpdateComment(r.Context(), comment)

	if err != nil {
		writeStoreError(w, err, "comment")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Store.DeleteComment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// attachments

type attachmentCreateRequest struct {
	TicketID int64  `json:"ticket_id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
}

type attachmentPatchRequest struct {
	TicketID *int64  `json:"ticket_id"`
	FileName *string `json:"file_name"`
	FileURL  *string `json:"file_url"`
}

func (h *Handler) GetAttachments(w http.ResponseWriter, r *http.Request) {
	attachments, err := h.Store.GetAttachments(r.Context())

	if err != nil {
		writeStoreError(w, err, "attachments")
		return
	}

	writeJSON(w, http.StatusOK, attachments)
}

func (h *Handler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	attachment, err := h.Store.GetAttachment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "attachment")
		return
	}

	writeJSON(w, http.StatusOK, attachment)
}

func (h *Handler) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	var req attachmentCreateRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.FileName = strings.TrimSpace(req.FileName)
	req.FileURL = strings.TrimSpace(req.FileURL)

	if req.TicketID <= 0 ||
		req.FileName == "" ||
		req.FileURL == "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid attachment data")
		return
	}

	attachment, err := h.Store.CreateAttachment(r.Context(), Attachment{
		TicketID: req.TicketID,
		FileName: req.FileName,
		FileURL:  req.FileURL,
	})

	if err != nil {
		writeStoreError(w, err, "attachment")
		return
	}

	writeJSON(w, http.StatusCreated, attachment)
}

func (h *Handler) UpdateAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req attachmentPatchRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	attachment, err := h.Store.GetAttachment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "attachment")
		return
	}

	if req.TicketID != nil {
		if *req.TicketID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "invalid ticket_id")
			return
		}

		attachment.TicketID = *req.TicketID
	}

	if req.FileName != nil {
		fileName := strings.TrimSpace(*req.FileName)

		if fileName == "" {
			writeError(w, http.StatusUnprocessableEntity, "file_name cannot be empty")
			return
		}

		attachment.FileName = fileName
	}

	if req.FileURL != nil {
		fileURL := strings.TrimSpace(*req.FileURL)

		if fileURL == "" {
			writeError(w, http.StatusUnprocessableEntity, "file_url cannot be empty")
			return
		}

		attachment.FileURL = fileURL
	}

	attachment, err = h.Store.UpdateAttachment(r.Context(), attachment)

	if err != nil {
		writeStoreError(w, err, "attachment")
		return
	}

	writeJSON(w, http.StatusOK, attachment)
}

func (h *Handler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.Store.DeleteAttachment(r.Context(), id)

	if err != nil {
		writeStoreError(w, err, "attachment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
