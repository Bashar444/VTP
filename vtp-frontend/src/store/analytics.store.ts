import { create } from 'zustand';
import type { Alert } from '@/types/api';

interface AnalyticsState {
  alerts: Alert[];
  isLoading: boolean;

  // Actions
  setAlerts: (alerts: Alert[]) => void;
  addAlert: (alert: Alert) => void;
  removeAlert: (alertId: string) => void;
  markAlertAsRead: (alertId: string) => void;
  setLoading: (loading: boolean) => void;
}

export const useAnalyticsStore = create<AnalyticsState>((set) => ({
  alerts: [],
  isLoading: false,

  setAlerts: (alerts: Alert[]) => set({ alerts }),
  addAlert: (alert: Alert) =>
    set((state) => ({
      alerts: [alert, ...state.alerts],
    })),
  removeAlert: (alertId: string) =>
    set((state) => ({
      alerts: state.alerts.filter((a) => a.id !== alertId),
    })),
  markAlertAsRead: (alertId: string) =>
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.id === alertId ? { ...a, read: true } : a
      ),
    })),
  setLoading: (loading: boolean) => set({ isLoading: loading }),
}));
