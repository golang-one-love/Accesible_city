from pydantic import BaseModel, Field
from typing import Optional, List
from uuid import UUID
from datetime import datetime
from enum import Enum


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
    features: List[AccessibilityFeatureStr] = []
    entrance_step_height_cm: Optional[int] = None
    door_width_cm: Optional[int] = None
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
    accessibility: Optional[AccessibilityProfileDTO] = None


class UpdatePOIRequest(BaseModel):
    name: Optional[str] = Field(default=None, min_length=1, max_length=200)
    category: Optional[POICategoryStr] = None
    address: Optional[str] = Field(default=None, max_length=500)
    phone: Optional[str] = Field(default=None, max_length=50)
    website: Optional[str] = Field(default=None, max_length=200)
    opening_hours: Optional[str] = Field(default=None, max_length=200)
    accessibility: Optional[AccessibilityProfileDTO] = None


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
    owner_id: Optional[UUID]
    is_verified: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class POIListResponse(BaseModel):
    pois: List[POIResponse]
    total: int
    limit: int
    offset: int


class SearchNearbyRequest(BaseModel):
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)
    radius_meters: float = Field(gt=0, le=50000)
    category: Optional[POICategoryStr] = None
    limit: int = Field(default=20, le=100)


class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None