export type POICategory = 
  | 'restaurant' | 'cafe' | 'shop' | 'pharmacy' | 'hospital' | 'clinic'
  | 'bank' | 'post_office' | 'government' | 'park' | 'museum' | 'theater'
  | 'library' | 'school' | 'university' | 'hotel' | 'transport' | 'other'

export type AccessibilityFeature = 
  | 'ramp' | 'elevator' | 'wide_door' | 'accessible_toilet' | 'tactile_paving'
  | 'braille_signs' | 'audio_guide' | 'low_counter' | 'parking' | 'induction_loop'

export interface AccessibilityProfile {
  features: AccessibilityFeature[]
  entrance_step_height_cm: number | null
  door_width_cm: number | null
  has_accessible_toilet: boolean
  notes: string
}

export interface POI {
  id: string
  name: string
  category: POICategory
  latitude: number
  longitude: number
  address: string
  phone: string
  website: string
  opening_hours: string
  accessibility: AccessibilityProfile
  owner_id: string | null
  is_verified: boolean
  created_at: string
  updated_at: string
}

export interface POIListResponse {
  pois: POI[]
  total: number
  limit: number
  offset: number
}

export interface CreatePOIRequest {
  name: string
  category: POICategory
  latitude: number
  longitude: number
  address: string
  phone?: string
  website?: string
  opening_hours?: string
  accessibility?: Partial<AccessibilityProfile>
}

export interface SearchNearbyRequest {
  latitude: number
  longitude: number
  radius_meters: number
  category?: POICategory
  limit?: number
}