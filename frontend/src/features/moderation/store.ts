import { create } from 'zustand'
import { ModerationRequest, ModerationStatus } from '@entities/moderation/types'

interface ModerationState {
  requests: ModerationRequest[]
  selectedRequest: ModerationRequest | null
  filters: { status?: ModerationStatus }
  setRequests: (requests: ModerationRequest[]) => void
  setSelectedRequest: (request: ModerationRequest | null) => void
  setFilters: (filters: Partial<{ status: ModerationStatus }>) => void
  updateRequest: (id: string, updates: Partial<ModerationRequest>) => void
}

export const useModerationStore = create<ModerationState>((set) => ({
  requests: [],
  selectedRequest: null,
  filters: {},
  
  setRequests: (requests) => set({ requests }),
  
  setSelectedRequest: (request) => set({ selectedRequest: request }),
  
  setFilters: (filters) => set((state) => ({ 
    filters: { ...state.filters, ...filters } 
  })),
  
  updateRequest: (id, updates) => set((state) => ({
    requests: state.requests.map((r) => r.id === id ? { ...r, ...updates } : r),
    selectedRequest: state.selectedRequest?.id === id 
      ? { ...state.selectedRequest, ...updates } 
      : state.selectedRequest,
  })),
}))