#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Indonesian Calendar Module
==========================
Event detection for Indonesian market.
Includes payday, twin dates, holidays, and e-commerce events.
Platform-agnostic - usable for TikTok, Shopee, Lazada, etc.
"""

from datetime import datetime, date, timedelta
from typing import Dict, List, Any, Optional, Union, Tuple
from enum import Enum
from dataclasses import dataclass


class EventType(Enum):
    """Types of Indonesian shopping events."""
    PAYDAY_PRIME = "PAYDAY_PRIME"       # Tanggal 25-28, 1-3
    POST_PAYDAY = "POST_PAYDAY"         # Tanggal 4-10
    MID_MONTH = "MID_MONTH"             # Tanggal 11-17
    DRY_SEASON = "DRY_SEASON"           # Tanggal 18-24
    TWIN_DATE = "TWIN_DATE"             # 1.1, 2.2, ..., 12.12
    NATIONAL_HOLIDAY = "NATIONAL_HOLIDAY"
    WEEKEND = "WEEKEND"
    NORMAL = "NORMAL"


@dataclass
class EventInfo:
    """Information about a calendar event."""
    eventType: EventType
    multiplier: float
    description: str


class IndonesianCalendar:
    """
    Indonesian shopping calendar engine.
    Provides event detection and multipliers for ad performance analysis.
    
    Key events:
    - Payday Prime: 25th-3rd (salary disbursement period)
    - Post Payday: 4th-10th
    - Mid Month: 11th-17th  
    - Dry Season: 18th-24th (waiting for salary)
    - Twin Date: 1.1, 2.2, ..., 12.12
    - National Holidays
    - Weekend
    """
    
    # National holidays 2025-2026 (Tanggal Merah)
    NATIONAL_HOLIDAYS = {
        # 2025
        date(2025, 1, 1): "Tahun Baru 2025",
        date(2025, 1, 29): "Isra Mi'raj",
        date(2025, 3, 29): "Hari Raya Nyepi",
        date(2025, 3, 30): "Idul Fitri 1446 H",
        date(2025, 3, 31): "Idul Fitri 1446 H",
        date(2025, 4, 18): "Jumat Agung",
        date(2025, 5, 1): "Hari Buruh",
        date(2025, 5, 12): "Hari Raya Waisak",
        date(2025, 5, 29): "Kenaikan Isa Almasih",
        date(2025, 6, 1): "Hari Lahir Pancasila",
        date(2025, 6, 6): "Idul Adha 1446 H",
        date(2025, 6, 27): "Tahun Baru Islam 1447 H",
        date(2025, 8, 17): "Hari Kemerdekaan RI",
        date(2025, 9, 5): "Maulid Nabi Muhammad SAW",
        date(2025, 12, 25): "Hari Natal",
        # 2026
        date(2026, 1, 1): "Tahun Baru 2026",
        date(2026, 1, 17): "Isra Mi'raj 1447 H",
        date(2026, 3, 19): "Idul Fitri 1447 H",
        date(2026, 3, 20): "Idul Fitri 1447 H",
    }
    
    # Payday periods
    PAYDAY_PRIME_DAYS = list(range(25, 32)) + list(range(1, 4))  # 25-31, 1-3
    POST_PAYDAY_DAYS = list(range(4, 11))   # 4-10
    MID_MONTH_DAYS = list(range(11, 18))    # 11-17
    DRY_SEASON_DAYS = list(range(18, 25))   # 18-24
    
    # Multipliers
    MULTIPLIERS = {
        EventType.TWIN_DATE: 1.50,
        EventType.PAYDAY_PRIME: 1.30,
        EventType.POST_PAYDAY: 1.15,
        EventType.NATIONAL_HOLIDAY: 1.10,
        EventType.WEEKEND: 1.10,
        EventType.MID_MONTH: 0.95,
        EventType.DRY_SEASON: 0.80,
        EventType.NORMAL: 1.00,
    }
    
    def _parse_date(self, dateInput: Union[date, datetime, str]) -> Tuple[date, int]:
        """Parse input to date object and return (date, weekday)."""
        if isinstance(dateInput, str):
            dt = datetime.strptime(dateInput[:10], "%Y-%m-%d")
            return dt.date(), dt.weekday()
        elif isinstance(dateInput, datetime):
            return dateInput.date(), dateInput.weekday()
        elif isinstance(dateInput, date):
            return dateInput, datetime.combine(dateInput, datetime.min.time()).weekday()
        return dateInput, 0
    
    def getEventInfo(self, dateInput: Union[date, datetime, str]) -> EventInfo:
        """
        Get event information for a specific date.
        Returns the most significant event (highest multiplier).
        """
        dateObj, dayOfWeek = self._parse_date(dateInput)
        day = dateObj.day
        month = dateObj.month
        
        events: List[EventInfo] = []
        
        # Check Twin Date (1.1, 2.2, ..., 12.12)
        if day == month:
            events.append(EventInfo(
                eventType=EventType.TWIN_DATE,
                multiplier=self.MULTIPLIERS[EventType.TWIN_DATE],
                description=f"Tanggal Kembar {month}.{day}"
            ))
        
        # Check National Holiday
        if dateObj in self.NATIONAL_HOLIDAYS:
            events.append(EventInfo(
                eventType=EventType.NATIONAL_HOLIDAY,
                multiplier=self.MULTIPLIERS[EventType.NATIONAL_HOLIDAY],
                description=self.NATIONAL_HOLIDAYS[dateObj]
            ))
        
        # Check Payday Period
        if day in self.PAYDAY_PRIME_DAYS:
            events.append(EventInfo(
                eventType=EventType.PAYDAY_PRIME,
                multiplier=self.MULTIPLIERS[EventType.PAYDAY_PRIME],
                description="Periode Gajian (Prime)"
            ))
        elif day in self.POST_PAYDAY_DAYS:
            events.append(EventInfo(
                eventType=EventType.POST_PAYDAY,
                multiplier=self.MULTIPLIERS[EventType.POST_PAYDAY],
                description="Pasca Gajian"
            ))
        elif day in self.MID_MONTH_DAYS:
            events.append(EventInfo(
                eventType=EventType.MID_MONTH,
                multiplier=self.MULTIPLIERS[EventType.MID_MONTH],
                description="Pertengahan Bulan"
            ))
        elif day in self.DRY_SEASON_DAYS:
            events.append(EventInfo(
                eventType=EventType.DRY_SEASON,
                multiplier=self.MULTIPLIERS[EventType.DRY_SEASON],
                description="Tanggal Tua (Nunggu Gajian)"
            ))
        
        # Check Weekend
        if dayOfWeek >= 5:  # Saturday=5, Sunday=6
            events.append(EventInfo(
                eventType=EventType.WEEKEND,
                multiplier=self.MULTIPLIERS[EventType.WEEKEND],
                description="Weekend"
            ))
        
        # Return highest multiplier event, or NORMAL if none
        if events:
            return max(events, key=lambda e: e.multiplier)
        
        return EventInfo(
            eventType=EventType.NORMAL,
            multiplier=self.MULTIPLIERS[EventType.NORMAL],
            description="Hari Normal"
        )
    
    def getMultiplier(self, dateInput: Union[date, datetime, str]) -> float:
        """Get the multiplier for a specific date."""
        return self.getEventInfo(dateInput).multiplier
    
    def getCompositeMultiplier(
        self, dateInput: Union[date, datetime, str]
    ) -> Tuple[float, List[str]]:
        """
        Get composite multiplier considering all applicable events.
        Returns (multiplier, list of event descriptions).
        """
        dateObj, dayOfWeek = self._parse_date(dateInput)
        day = dateObj.day
        month = dateObj.month
        
        multiplier = 1.0
        descriptions = []
        
        # Twin Date
        if day == month:
            multiplier *= self.MULTIPLIERS[EventType.TWIN_DATE]
            descriptions.append(f"Tanggal Kembar {month}.{day}")
        
        # National Holiday
        if dateObj in self.NATIONAL_HOLIDAYS:
            multiplier *= 1.05  # Smaller boost if combined
            descriptions.append(self.NATIONAL_HOLIDAYS[dateObj])
        
        # Payday Effect
        if day in self.PAYDAY_PRIME_DAYS:
            multiplier *= self.MULTIPLIERS[EventType.PAYDAY_PRIME]
            descriptions.append("Periode Gajian")
        elif day in self.DRY_SEASON_DAYS:
            multiplier *= self.MULTIPLIERS[EventType.DRY_SEASON]
            descriptions.append("Tanggal Tua")
        
        # Weekend
        if dayOfWeek >= 5:
            multiplier *= 1.05  # Smaller boost if combined
            descriptions.append("Weekend")
        
        if not descriptions:
            descriptions.append("Normal")
        
        return (round(multiplier, 2), descriptions)
    
    def get_upcoming_events(
        self,
        start_date: Union[date, datetime, str, None] = None,
        days_ahead: int = 30
    ) -> List[Dict[str, Any]]:
        """Get upcoming significant events within date range."""
        if start_date is None:
            start_date = date.today()
        
        start, _ = self._parse_date(start_date)
        events = []
        
        for i in range(days_ahead):
            d = start + timedelta(days=i)
            event_info = self.getEventInfo(d)
            
            # Only include significant events (multiplier > 1.0 or < 1.0)
            if event_info.eventType not in [EventType.NORMAL, EventType.MID_MONTH]:
                events.append({
                    "date": d.isoformat(),
                    "type": event_info.eventType.value,
                    "description": event_info.description,
                    "multiplier": event_info.multiplier
                })
        
        return events


# ============================================================================
# Helper Functions
# ============================================================================

def is_payday(d: Union[date, datetime, str]) -> bool:
    """Check if date is in payday prime period (25th-3rd)."""
    cal = IndonesianCalendar()
    dateObj, _ = cal._parse_date(d)
    return dateObj.day in IndonesianCalendar.PAYDAY_PRIME_DAYS


def is_twin_date(d: Union[date, datetime, str]) -> bool:
    """Check if date is a twin date (1.1, 2.2, etc.)."""
    cal = IndonesianCalendar()
    dateObj, _ = cal._parse_date(d)
    return dateObj.day == dateObj.month


def is_holiday(d: Union[date, datetime, str]) -> bool:
    """Check if date is a national holiday."""
    cal = IndonesianCalendar()
    dateObj, _ = cal._parse_date(d)
    return dateObj in IndonesianCalendar.NATIONAL_HOLIDAYS


def get_event_boost(d: Union[date, datetime, str]) -> float:
    """Get boost multiplier for date."""
    cal = IndonesianCalendar()
    return cal.getMultiplier(d)


def get_upcoming_events(
    start_date: Union[date, datetime, str, None] = None,
    days_ahead: int = 30
) -> List[Dict[str, Any]]:
    """Get upcoming events within date range."""
    cal = IndonesianCalendar()
    return cal.get_upcoming_events(start_date, days_ahead)
