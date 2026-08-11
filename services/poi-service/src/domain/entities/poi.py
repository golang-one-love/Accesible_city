from enum import Enum
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, List, Dict, Any
from uuid import UUID, uuid4


class POICategory(str, Enum):
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


class AccessibilityFeature(str, Enum):
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


@dataclass
class AccessibilityProfile:
    features: List[AccessibilityFeature] = field(default_factory=list)
    entrance_step_height_cm: Optional[int] = None
    door_width_cm: Optional[int] = None
    has_accessible_toilet: bool = False
    notes: str = ""

    def add_feature(self, feature: AccessibilityFeature) -> None:
        if feature not in self.features:
            self.features.append(feature)

    def remove_feature(self, feature: AccessibilityFeature) -> None:
        if feature in self.features:
            self.features.remove(feature)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "features": [f.value for f in self.features],
            "entrance_step_height_cm": self.entrance_step_height_cm,
            "door_width_cm": self.door_width_cm,
            "has_accessible_toilet": self.has_accessible_toilet,
            "notes": self.notes,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "AccessibilityProfile":
        return cls(
            features=[AccessibilityFeature(f) for f in data.get("features", [])],
            entrance_step_height_cm=data.get("entrance_step_height_cm"),
            door_width_cm=data.get("door_width_cm"),
            has_accessible_toilet=data.get("has_accessible_toilet", False),
            notes=data.get("notes", ""),
        )


@dataclass
class PointOfInterest:
    id: UUID = field(default_factory=uuid4)
    name: str = ""
    category: POICategory = POICategory.OTHER
    latitude: float = 0.0
    longitude: float = 0.0
    address: str = ""
    phone: str = ""
    website: str = ""
    opening_hours: str = ""
    accessibility: AccessibilityProfile = field(default_factory=AccessibilityProfile)
    owner_id: Optional[UUID] = None
    is_verified: bool = False
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)

    def update_accessibility(self, profile: AccessibilityProfile) -> None:
        self.accessibility = profile
        self.updated_at = datetime.utcnow()

    def verify(self) -> None:
        self.is_verified = True
        self.updated_at = datetime.utcnow()