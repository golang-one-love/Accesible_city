import { create } from 'zustand'
import { Barrier, BarrierType, BarrierStatus, Severity, Coordinates } from '@entities/barrier/types'

interface BarrierFilters {
  status?: BarrierStatus
  type?: BarrierType
  severity_min?: Severity
  severity_max?: Severity
  bounds?: [Coordinates, Coordinates]
}

interface BarrierState {
  barriers: Barrier[]
  selectedBarrier: Barrier | null
  filters: BarrierFilters
  setBarriers: (barriers: Barrier[]) => void
  setSelectedBarrier: (barrier: Barrier | null) => void
  setFilters: (filters: Partial<BarrierFilters>) => void
  clearFilters: () => void
  addBarrier: (barrier: Barrier) => void
  updateBarrier: (id: string, updates: Partial<Barrier>) => void
  removeBarrier: (id: string) => void
}

export const useBarrierStore = create<BarrierState>((set) => ({
  barriers: [],
  selectedBarrier: null,
  filters: {},
  
  setBarriers: (barriers) => set({ barriers }),
  
  setSelectedBarrier: (barrier) => set({ selectedBarrier: barrier }),
  
  setFilters: (filters) => set((state) => ({ 
    filters: { ...state.filters, ...filters } 
  })),
  
  clearFilters: () => set({ filters: {} }),
  
  addBarrier: (barrier) => set((state) => ({ 
    barriers: [barrier, ...state.barriers] 
  })),
  
  updateBarrier: (id, updates) => set((state) => ({
    barriers: state.barriers.map((b) => b.id === id ? { ...b, ...updates } : b),
    selectedBarrier: state.selectedBarrier?.id === id 
      ? { ...state.selectedBarrier, ...updates } 
      : state.selectedBarrier,
  })),
  
  removeBarrier: (id) => set((state) => ({
    barriers: state.barriers.filter((b) => b.id !== id),
    selectedBarrier: state.selectedBarrier?.id === id ? null : state.selectedBarrier,
  })),
}))