
import pytest
from src.domain.entities.poi import (
    AccessibilityFeature,
    AccessibilityProfile,
    POICategory,
    PointOfInterest,
)


@pytest.fixture
def poi() -> PointOfInterest:
    return PointOfInterest(
        name="Test cafe",
        category=POICategory.CAFE,
        latitude=55.75,
        longitude=37.61,
        address="Test street 1",
    )


def test_create_poi_defaults(poi: PointOfInterest) -> None:
    assert not poi.is_verified
    assert poi.accessibility.features == []


def test_add_feature(poi: PointOfInterest) -> None:
    poi.accessibility.add_feature(AccessibilityFeature.RAMP)
    assert AccessibilityFeature.RAMP in poi.accessibility.features


def test_add_feature_is_idempotent(poi: PointOfInterest) -> None:
    poi.accessibility.add_feature(AccessibilityFeature.RAMP)
    poi.accessibility.add_feature(AccessibilityFeature.RAMP)
    assert poi.accessibility.features.count(AccessibilityFeature.RAMP) == 1


def test_remove_feature(poi: PointOfInterest) -> None:
    poi.accessibility.add_feature(AccessibilityFeature.RAMP)
    poi.accessibility.remove_feature(AccessibilityFeature.RAMP)
    assert poi.accessibility.features == []


def test_verify_sets_verified_flag(poi: PointOfInterest) -> None:
    poi.verify()
    assert poi.is_verified
    assert poi.updated_at is not None


def test_update_accessibility(poi: PointOfInterest) -> None:
    profile = AccessibilityProfile(
        features=[AccessibilityFeature.INDUCTION_LOOP],
        entrance_step_height_cm=5,
        notes="Main entrance only",
    )
    poi.update_accessibility(profile)
    assert poi.accessibility == profile
    assert poi.updated_at is not None


def test_accessibility_profile_to_dict() -> None:
    profile = AccessibilityProfile(features=[AccessibilityFeature.ELEVATOR], notes="n")
    data = profile.to_dict()
    assert data["features"] == ["elevator"]
    assert data["notes"] == "n"
