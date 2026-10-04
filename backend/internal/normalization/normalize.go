package normalization

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"transjakarta/internal/domain"
)

// GenerateStopID creates a deterministic stop identifier based on stop name and coordinates
func GenerateStopID(stopName string, lat, lon float64) string {
	cleanName := strings.TrimSpace(strings.ToUpper(stopName))
	payload := fmt.Sprintf("%s|%.5f|%.5f", cleanName, lat, lon)
	hash := sha256.Sum256([]byte(payload))
	hashHex := hex.EncodeToString(hash[:])[:8]

	slug := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, cleanName)

	if len(slug) > 16 {
		slug = slug[:16]
	}
	slug = strings.Trim(slug, "_")

	return fmt.Sprintf("STP-%s-%s", slug, hashHex)
}

// ClassifyService identifies service type (BRT, Feeder, Mikrotrans, Amari)
func ClassifyService(routeID string) domain.ServiceType {
	cleanID := strings.TrimSpace(routeID)
	if _, err := strconv.Atoi(cleanID); err == nil {
		return domain.ServiceBRT
	}
	if strings.HasPrefix(cleanID, "JAK.") || strings.HasPrefix(cleanID, "JAK") {
		return domain.ServiceMikrotrans
	}
	if strings.HasPrefix(cleanID, "M") {
		return domain.ServiceAmari
	}
	return domain.ServiceNonBRT
}
