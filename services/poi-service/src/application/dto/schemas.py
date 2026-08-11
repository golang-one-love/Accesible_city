from datetime import datetime
from enum import Enum
from uuid import UUID

from pydantic import BaseModel, Field


class POICategoryStr(str, Enum):
    RESTAURANT = "restaurant"
    CAFE = "cafe"
    SHOP = "shop"
    PHARMACY = "pharmacy"
    HOSPITAL = "hospital"
    CLINIC = "clinic"
    BANK = "bank"
    POST_OFFICE = "post_office"
    GOVERNMENT = "government"
    PARK = "park"
    MUSEUM = "museum"
    THEATER = "theater"
    LIBRARY = "library"
    SCHOOL = "school"
    UNIVERSITY = "university"
    HOTEL = "hotel"
    TRANSPORT = "transport"
    OTHER = "other"


class AccessibilityFeatureStr(str, Enum):
    RAMP = "ramp"
    ELEVATOR = "elevator"
    WIDE_DOOR = "wide_door"
    ACCESSIBLE_TOILET = "accessible_toilet"
    TACTILE_PAVING = "tactile_paving"
    BRAILLE_SIGNS = "braille_signs"
    AUDIO_GUIDE = "audio_guide"
    LOW_COUNTER = "low_counter"
    PARKING = "parking"
    INDUCTION_LOOP = "induction_loop"


class AccessibilityProfileDTO(BaseModel):
    features: list[AccessibilityFeatureStr] = []
    entrance_step_height_cm: int | None = None
    door_width_cm: int | None = None
    has_accessible_toilet: bool = False
    notes: str = ""


class CreatePOIRequest(BaseModel):
    name: str = Field(min_length=1, max_length=200)
    category: POICategoryStr
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)
    address: str = Field(min_length=1, max_length=500)
    phone: str = Field(default="", max_length=50)
    website: str = Field(default="", max_length=200)
    opening_hours: str = Field(default="", max_length=200)
    accessibility: AccessibilityProfileDTO | None = None


class UpdatePOIRequest(BaseModel):
    name: str | None = Field(default=None, min_length=1, max_length=200)
    category: POICategoryStr | None = None
    address: str | None = Field(default=None, max_length=500)
    phone: str | None = Field(default=None, max_length=50)
    website: str | None = Field(default=None, max_length=200)
    opening_hours: str | None = Field(default=None, max_length=200)
    accessibility: AccessibilityProfileDTO | None = None


class POIResponse(BaseModel):
    id: UUID
    name: str
    category: POICategoryStr
    latitude: float
    longitude: float
    address: str
    phone: str
    website: str
    opening_hours: str
    accessibility: AccessibilityProfileDTO
    owner_id: UUID | None
    is_verified: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class POIListResponse(BaseModel):
    pois: list[POIResponse]
    total: int
    limit: int
    offset: int


class SearchNearbyRequest(BaseModel):
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)
    radius_meters: float = Field(gt=0, le=50000)
    category: POICategoryStr | None = None
    limit: int = Field(default=20, le=100)


class ErrorResponse(BaseModel):
    error: str
    detail: str | None = None