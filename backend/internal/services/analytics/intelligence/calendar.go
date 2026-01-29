package intelligence

import (
	"time"
)

// EventType represents types of Indonesian shopping events
type EventType string

const (
	EventTypePaydayPrime     EventType = "PAYDAY_PRIME"
	EventTypePostPayday      EventType = "POST_PAYDAY"
	EventTypeMidMonth        EventType = "MID_MONTH"
	EventTypeDrySeason       EventType = "DRY_SEASON"
	EventTypeTwinDate        EventType = "TWIN_DATE"
	EventTypeNationalHoliday EventType = "NATIONAL_HOLIDAY"
	EventTypeWeekend         EventType = "WEEKEND"
	EventTypeNormal          EventType = "NORMAL"
)

// EventInfo contains information about a calendar event
type EventInfo struct {
	EventType   EventType `json:"event_type"`
	Multiplier  float64   `json:"multiplier"`
	Description string    `json:"description"`
}

// IndonesianCalendar provides event detection for Indonesian market
type IndonesianCalendar struct {
	holidays map[string]string
}

// NewIndonesianCalendar creates a new calendar instance
func NewIndonesianCalendar() *IndonesianCalendar {
	cal := &IndonesianCalendar{
		holidays: make(map[string]string),
	}
	cal.initHolidays()
	return cal
}

// initHolidays initializes Indonesian national holidays
func (c *IndonesianCalendar) initHolidays() {
	// 2025 Holidays
	c.holidays["2025-01-01"] = "Tahun Baru 2025"
	c.holidays["2025-01-29"] = "Isra Mi'raj"
	c.holidays["2025-03-29"] = "Hari Raya Nyepi"
	c.holidays["2025-03-30"] = "Idul Fitri 1446 H"
	c.holidays["2025-03-31"] = "Idul Fitri 1446 H"
	c.holidays["2025-04-18"] = "Jumat Agung"
	c.holidays["2025-05-01"] = "Hari Buruh"
	c.holidays["2025-05-12"] = "Hari Raya Waisak"
	c.holidays["2025-05-29"] = "Kenaikan Isa Almasih"
	c.holidays["2025-06-01"] = "Hari Lahir Pancasila"
	c.holidays["2025-06-06"] = "Idul Adha 1446 H"
	c.holidays["2025-06-27"] = "Tahun Baru Islam 1447 H"
	c.holidays["2025-08-17"] = "Hari Kemerdekaan RI"
	c.holidays["2025-09-05"] = "Maulid Nabi Muhammad SAW"
	c.holidays["2025-12-25"] = "Hari Natal"
	// 2026 Holidays
	c.holidays["2026-01-01"] = "Tahun Baru 2026"
	c.holidays["2026-01-17"] = "Isra Mi'raj 1447 H"
	c.holidays["2026-03-19"] = "Idul Fitri 1447 H"
	c.holidays["2026-03-20"] = "Idul Fitri 1447 H"
}

// Multipliers for each event type
var eventMultipliers = map[EventType]float64{
	EventTypeTwinDate:        1.50,
	EventTypePaydayPrime:     1.30,
	EventTypePostPayday:      1.15,
	EventTypeNationalHoliday: 1.10,
	EventTypeWeekend:         1.10,
	EventTypeMidMonth:        0.95,
	EventTypeDrySeason:       0.80,
	EventTypeNormal:          1.00,
}

// GetEventInfo returns event information for a specific date
func (c *IndonesianCalendar) GetEventInfo(date time.Time) EventInfo {
	day := date.Day()
	month := int(date.Month())
	weekday := date.Weekday()
	dateStr := date.Format("2006-01-02")

	var events []EventInfo

	// Check Twin Date (1.1, 2.2, ..., 12.12)
	if day == month {
		events = append(events, EventInfo{
			EventType:   EventTypeTwinDate,
			Multiplier:  eventMultipliers[EventTypeTwinDate],
			Description: "Tanggal Kembar",
		})
	}

	// Check National Holiday
	if holidayName, exists := c.holidays[dateStr]; exists {
		events = append(events, EventInfo{
			EventType:   EventTypeNationalHoliday,
			Multiplier:  eventMultipliers[EventTypeNationalHoliday],
			Description: holidayName,
		})
	}

	// Check Payday Period (25th-31st, 1st-3rd)
	if (day >= 25 && day <= 31) || (day >= 1 && day <= 3) {
		events = append(events, EventInfo{
			EventType:   EventTypePaydayPrime,
			Multiplier:  eventMultipliers[EventTypePaydayPrime],
			Description: "Periode Gajian (Prime)",
		})
	} else if day >= 4 && day <= 10 {
		events = append(events, EventInfo{
			EventType:   EventTypePostPayday,
			Multiplier:  eventMultipliers[EventTypePostPayday],
			Description: "Pasca Gajian",
		})
	} else if day >= 11 && day <= 17 {
		events = append(events, EventInfo{
			EventType:   EventTypeMidMonth,
			Multiplier:  eventMultipliers[EventTypeMidMonth],
			Description: "Pertengahan Bulan",
		})
	} else if day >= 18 && day <= 24 {
		events = append(events, EventInfo{
			EventType:   EventTypeDrySeason,
			Multiplier:  eventMultipliers[EventTypeDrySeason],
			Description: "Tanggal Tua (Nunggu Gajian)",
		})
	}

	// Check Weekend
	if weekday == time.Saturday || weekday == time.Sunday {
		events = append(events, EventInfo{
			EventType:   EventTypeWeekend,
			Multiplier:  eventMultipliers[EventTypeWeekend],
			Description: "Weekend",
		})
	}

	// Return highest multiplier event
	if len(events) > 0 {
		best := events[0]
		for _, e := range events[1:] {
			if e.Multiplier > best.Multiplier {
				best = e
			}
		}
		return best
	}

	return EventInfo{
		EventType:   EventTypeNormal,
		Multiplier:  eventMultipliers[EventTypeNormal],
		Description: "Hari Normal",
	}
}

// GetMultiplier returns the multiplier for a specific date
func (c *IndonesianCalendar) GetMultiplier(date time.Time) float64 {
	return c.GetEventInfo(date).Multiplier
}

// GetUpcomingEvents returns upcoming significant events
func (c *IndonesianCalendar) GetUpcomingEvents(startDate time.Time, daysAhead int) []map[string]interface{} {
	var events []map[string]interface{}

	for i := 0; i < daysAhead; i++ {
		d := startDate.AddDate(0, 0, i)
		info := c.GetEventInfo(d)

		// Only include significant events
		if info.EventType != EventTypeNormal && info.EventType != EventTypeMidMonth {
			events = append(events, map[string]interface{}{
				"date":        d.Format("2006-01-02"),
				"type":        string(info.EventType),
				"description": info.Description,
				"multiplier":  info.Multiplier,
			})
		}
	}

	return events
}

// IsPayday checks if date is in payday prime period
func IsPayday(date time.Time) bool {
	day := date.Day()
	return (day >= 25 && day <= 31) || (day >= 1 && day <= 3)
}

// IsTwinDate checks if date is a twin date
func IsTwinDate(date time.Time) bool {
	return date.Day() == int(date.Month())
}
