package service

import "time"

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type FixedClock struct{ Time time.Time }

func (c FixedClock) Now() time.Time { return c.Time }
