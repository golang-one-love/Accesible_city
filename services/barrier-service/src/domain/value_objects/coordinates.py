from dataclasses import dataclass
from typing import Tuple


@dataclass(frozen=True)
class Coordinates:
    latitude: float
    longitude: float

    def __post_init__(self):
        if not -90 <= self.latitude <= 90:
            raise ValueError("Latitude must be between -90 and 90")
        if not -180 <= self.longitude <= 180:
            raise ValueError("Longitude must be between -180 and 180")

    def to_tuple(self) -> Tuple[float, float]:
        return (self.latitude, self.longitude)

    def distance_to(self, other: "Coordinates") -> float:
        from math import radians, sin, cos, sqrt, atan2
        R = 6371000  # Earth radius in meters

        lat1, lon1 = radians(self.latitude), radians(self.longitude)
        lat2, lon2 = radians(other.latitude), radians(other.longitude)

        dlat = lat2 - lat1
        dlon = lon2 - lon1

        a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
        c = 2 * atan2(sqrt(a), sqrt(1-a))

        return R * c

    @classmethod
    def from_tuple(cls, coords: Tuple[float, float]) -> "Coordinates":
        return cls(latitude=coords[0], longitude=coords[1])

    def __str__(self) -> str:
        return f"{self.latitude},{self.longitude}"