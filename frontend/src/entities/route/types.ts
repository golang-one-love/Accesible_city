export type MobilityProfile = 'wheelchair' | 'stroller' | 'elderly' | 'default'

export interface Coordinates {
  latitude: number
  longitude: number
}

export interface RouteNode {
  id: string
  latitude: number
  longitude: number
}

export interface RouteResponse {
  nodes: RouteNode[]
  total_distance: number
  max_severity: number
}

export interface BuildRouteRequest {
  start: Coordinates
  finish: Coordinates
  mobility_profile: MobilityProfile
}