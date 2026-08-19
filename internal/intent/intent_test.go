package intent

import (
	"testing"
)

func TestIntentValidation(t *testing.T) {
	reqInvalid := IntentRequest{
		Type:       "unknown_intent",
		Parameters: map[string]interface{}{},
	}

	_, err := ExecuteIntent(reqInvalid)
	if err == nil {
		t.Errorf("Se esperaba error con intención no soportada, se obtuvo nil")
	}
}
