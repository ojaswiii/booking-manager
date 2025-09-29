package controllers

import (
	"encoding/json"
	"net/http"

	"ticket-booking-system/src/internal/usecase"
	"ticket-booking-system/src/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EventController struct {
	eventUsecase *usecase.EventUsecase
	logger       *utils.Logger
}

// NewEventController creates a new event controller
func NewEventController(eventUsecase *usecase.EventUsecase, logger *utils.Logger) *EventController {
	return &EventController{
		eventUsecase: eventUsecase,
		logger:       logger,
	}
}

// CreateEvent handles POST /api/events
func (c *EventController) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := c.eventUsecase.CreateEvent(r.Context(), req)
	if err != nil {
		c.logger.Error("Failed to create event", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	c.respondWithJSON(w, http.StatusCreated, response)
}

// GetEvent handles GET /api/events/{id}
func (c *EventController) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := c.eventUsecase.GetEvent(r.Context(), eventID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.respondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		c.logger.Error("Failed to get event", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get event")
		return
	}

	c.respondWithJSON(w, http.StatusOK, event)
}

// GetAllEvents handles GET /api/events
func (c *EventController) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := c.eventUsecase.GetAllEvents(r.Context())
	if err != nil {
		c.logger.Error("Failed to get events", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	c.respondWithJSON(w, http.StatusOK, events)
}

// GetEventTickets handles GET /api/events/{id}/tickets
func (c *EventController) GetEventTickets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	tickets, err := c.eventUsecase.GetEventTickets(r.Context(), eventID)
	if err != nil {
		c.logger.Error("Failed to get event tickets", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get event tickets")
		return
	}

	c.respondWithJSON(w, http.StatusOK, tickets)
}

// GetAvailableTickets handles GET /api/events/{id}/tickets/available
func (c *EventController) GetAvailableTickets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := uuid.Parse(vars["id"])
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	tickets, err := c.eventUsecase.GetAvailableTickets(r.Context(), eventID)
	if err != nil {
		c.logger.Error("Failed to get available tickets", "error", err)
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get available tickets")
		return
	}

	c.respondWithJSON(w, http.StatusOK, tickets)
}

// Helper methods

func (c *EventController) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (c *EventController) respondWithError(w http.ResponseWriter, code int, message string) {
	c.respondWithJSON(w, code, map[string]string{"error": message})
}
