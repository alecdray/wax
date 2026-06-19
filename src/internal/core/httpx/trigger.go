package httpx

import (
	"encoding/json"
	"net/http"
)

// SetHXTrigger sets the HX-Trigger response header so HTMX fires a client-side
// event with the given name and detail payload. Call it before the response
// body is written. detail is serialised as the event's `detail`.
func SetHXTrigger(w http.ResponseWriter, event string, detail any) error {
	payload, err := json.Marshal(map[string]any{event: detail})
	if err != nil {
		return err
	}
	w.Header().Set("HX-Trigger", string(payload))
	return nil
}
