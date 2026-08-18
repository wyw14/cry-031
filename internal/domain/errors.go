package domain

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrInvalidState       = errors.New("invalid state transition")
	ErrCapacityExceeded   = errors.New("activity capacity exceeded")
	ErrScheduleConflict   = errors.New("volunteer schedule conflict")
	ErrActivityClosed     = errors.New("activity is not accepting claims")
	ErrInvalidDuration    = errors.New("service duration is invalid")
	ErrImmutableRecord    = errors.New("confirmed service records require a correction")
	ErrInvalidCorrection  = errors.New("correction is invalid")
	ErrUnacknowledgedRisk = errors.New("handoff still requires acknowledgement")
	ErrInvalidFilter      = errors.New("unsupported filter or sort value")
)
