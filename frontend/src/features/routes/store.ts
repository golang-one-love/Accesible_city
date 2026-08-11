import { create } from 'zustand'
import { Coordinates, RouteResponse, MobilityProfile } from '@entities/route/types'

interface RouteState {
  start: Coordinates | null
  finish: Coordinates | null
  profile: MobilityProfile
  routeResult: RouteResponse | null
  isBuilding: boolean
  setStart: (coords: Coordinates | null) => void
  setFinish: (coords: Coordinates | null) => void
  setProfile: (profile: MobilityProfile) => void
  setRouteResult: (result: RouteResponse | null) => void
  setIsBuilding: (building: boolean) => void
  clearRoute: () => void
}

export const useRouteStore = create<RouteState>((set) => ({
  start: null,
  finish: null,
  profile: 'default',
  routeResult: null,
  isBuilding: false,
  
  setStart: (coords) => set({ start: coords }),
  
  setFinish: (coords) => set({ finish: coords }),
  
  setProfile: (profile) => set({ profile }),
  
  setRouteResult: (result) => set({ routeResult: result }),
  
  setIsBuilding: (building) => set({ isBuilding: building }),
  
  clearRoute: () => set({ 
    start: null, 
    finish: null, 
    routeResult: null, 
    isBuilding: false 
  }),
}))