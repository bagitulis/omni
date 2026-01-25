#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Calendar Subpackage
===================
Indonesian calendar and event detection.
Platform-agnostic.
"""

from .indonesian import (
    IndonesianCalendar,
    EventType,
    EventInfo,
    is_payday,
    is_twin_date,
    is_holiday,
    get_event_boost,
    get_upcoming_events,
)

__all__ = [
    "IndonesianCalendar",
    "EventType",
    "EventInfo",
    "is_payday",
    "is_twin_date",
    "is_holiday",
    "get_event_boost",
    "get_upcoming_events",
]
