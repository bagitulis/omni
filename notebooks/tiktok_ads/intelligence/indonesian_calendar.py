#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Indonesian Calendar (Re-exported from Core)
===========================================
This module re-exports the Indonesian calendar from the shared core library.
All calendar logic is in `notebooks/core/calendar/indonesian.py`.

For backward compatibility, this module provides the same interface.
"""

import sys
from pathlib import Path

# Add parent directory to sys.path for core imports
_parent_dir = Path(__file__).resolve().parent.parent.parent
if str(_parent_dir) not in sys.path:
    sys.path.insert(0, str(_parent_dir))

# Re-export everything from core
from core.calendar.indonesian import (
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
