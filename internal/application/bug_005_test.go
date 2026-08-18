package application

import (
	"context"
	"testing"
	"time"
)

func TestCreateSlotRejectsBlankName(t *testing.T) {
	e, _, _ := fixture()
	_, err := e.CreateSlot(context.Background(), captain(), RequestMeta{}, CreateSlotRequest{ActivityID: "activity-visit", Name: "   ", Description: "x", Capacity: 1})
	if err == nil {
		t.Fatal("blank slot name accepted")
	}
	_ = time.Time{}
}
