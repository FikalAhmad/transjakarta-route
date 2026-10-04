package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"transjakarta/internal/domain"
	"transjakarta/internal/routing/graph"
	"transjakarta/internal/validation"
)

type APIHandler struct {
	Engine *graph.RoutingEngine
}

func NewAPIHandler(e *graph.RoutingEngine) *APIHandler {
	return &APIHandler{Engine: e}
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = msg
	writeJSON(w, status, resp)
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "UP",
	})
}

func (h *APIHandler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	if h.Engine == nil || h.Engine.Graph == nil || len(h.Engine.Graph.StopsByID) == 0 {
		writeError(w, http.StatusServiceUnavailable, "NOT_READY", "Routing graph not loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "READY",
		"total_stops": len(h.Engine.Graph.StopsByID),
		"routes":      len(h.Engine.Graph.RoutesByID),
	})
}

func (h *APIHandler) GetNearbyStops(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("latitude")
	lonStr := r.URL.Query().Get("longitude")
	radStr := r.URL.Query().Get("radius")
	limStr := r.URL.Query().Get("limit")

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "INVALID_COORDINATES", "latitude and longitude query parameters required")
		return
	}

	radius := 1000.0
	if rVal, err := strconv.ParseFloat(radStr, 64); err == nil && rVal > 0 {
		radius = rVal
	}

	limit := 10
	if lVal, err := strconv.Atoi(limStr); err == nil && lVal > 0 {
		limit = lVal
	}

	stops := h.Engine.FindCandidateStops(lat, lon, limit, radius)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"stops": stops,
	})
}

func (h *APIHandler) SearchRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST method required")
		return
	}

	var req domain.RouteSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON request payload")
		return
	}

	if err := validation.ValidateCoordinates(req.Origin.Latitude, req.Origin.Longitude); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ORIGIN", err.Error())
		return
	}
	if err := validation.ValidateCoordinates(req.Destination.Latitude, req.Destination.Longitude); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DESTINATION", err.Error())
		return
	}

	routes, err := h.Engine.SearchRoutes(req)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_ROUTE_FOUND", err.Error())
		return
	}

	// Validate each plan before sending response
	var validRoutes []domain.RoutePlanResult
	for _, plan := range routes {
		if err := validation.ValidateRoutePlan(&plan); err == nil {
			validRoutes = append(validRoutes, plan)
		}
	}

	if len(validRoutes) == 0 {
		writeError(w, http.StatusNotFound, "NO_VALID_ROUTE", "No structurally valid route found")
		return
	}

	resp := domain.RouteSearchResponse{}
	resp.Data.Routes = validRoutes
	writeJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	var routes []domain.Route
	for _, rObj := range h.Engine.Graph.RoutesByID {
		routes = append(routes, rObj)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"routes": routes,
	})
}
