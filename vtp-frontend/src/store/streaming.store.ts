import { create } from 'zustand';
import type { StreamingSession } from '@/types/api';

interface StreamingState {
  session: StreamingSession | null;
  isLive: boolean;
  participantList: string[];
  isLoading: boolean;

  // Actions
  setSession: (session: StreamingSession | null) => void;
  setIsLive: (isLive: boolean) => void;
  setParticipantList: (participants: string[]) => void;
  addParticipant: (participantId: string) => void;
  removeParticipant: (participantId: string) => void;
  setLoading: (loading: boolean) => void;
}

export const useStreamingStore = create<StreamingState>((set) => ({
  session: null,
  isLive: false,
  participantList: [],
  isLoading: false,

  setSession: (session: StreamingSession | null) => set({ session }),
  setIsLive: (isLive: boolean) => set({ isLive }),
  setParticipantList: (participants: string[]) =>
    set({ participantList: participants }),
  addParticipant: (participantId: string) =>
    set((state) => ({
      participantList: [...state.participantList, participantId],
    })),
  removeParticipant: (participantId: string) =>
    set((state) => ({
      participantList: state.participantList.filter((id) => id !== participantId),
    })),
  setLoading: (loading: boolean) => set({ isLoading: loading }),
}));
