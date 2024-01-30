package event_service

import (
	"encoding/json"

	"github.com/yapkah/go-api/helpers"
)

// EventSetup struct
type EventSetup struct {
	Quota int
}

// MapEventSetup func
func MapEventSetup(rawEventSetupData string) (EventSetup, string) {
	var eventSetup EventSetup

	// RawEventSetup struct
	type RawEventSetup struct {
		Quota string `json:"quota"`
	}

	// mapping event setting into struct
	rawEventSetup := &RawEventSetup{}
	err := json.Unmarshal([]byte(rawEventSetupData), rawEventSetup)
	if err != nil {
		return EventSetup{}, "Unmarshal" + err.Error()
	}

	// after map convert to relative type
	quota, err := helpers.ValueToInt(rawEventSetup.Quota)
	if err != nil {
		return EventSetup{}, "ValueToInt():1" + err.Error()
	}
	eventSetup.Quota = quota

	// for newly added setting that will not declare in every event please do a checking before convert to other type
	// if rawEventSetup.TestExtra != "" {
	// 	testExtra, err := helpers.ValueToInt(rawEventSetup.TestExtra)
	// 	if err != nil {
	// 		return EventSetup{}, "ValueToInt():2" + err.Error()
	// 	}
	// 	eventSetup.TestExtra = testExtra
	// }

	return eventSetup, ""
}
