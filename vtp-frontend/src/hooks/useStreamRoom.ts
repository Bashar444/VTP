"use client";
import { useEffect, useState, useCallback } from 'react';
import SignalingService from '@/services/signaling.service';
import { useAuthStore } from '@/store/auth.store';

interface StreamPeer {
  id: string;
  name?: string;
  role?: string;
  joinedAt: Date;
}

export function useStreamRoom(roomId: string) {
  const auth = useAuthStore();
  const [signaling, setSignaling] = useState<SignalingService | null>(null);
  const [peers, setPeers] = useState<StreamPeer[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (!auth.user || !roomId) return;
    const svc = new SignalingService(roomId, auth.user.id, auth.token || '');
    setSignaling(svc);

    svc.onPeerJoined((peerId, peerInfo) => {
      setPeers(prev => {
        if (prev.find(p => p.id === peerId)) return prev;
        return [...prev, { id: peerId, name: peerInfo?.name, role: peerInfo?.role, joinedAt: new Date() }];
      });
    });
    svc.onPeerLeft(peerId => {
      setPeers(prev => prev.filter(p => p.id !== peerId));
    });
    svc.onNewProducer((producerId, peerId, kind) => {
      // Placeholder: actual media consumption handled by useMediasoup
      console.log('New producer', producerId, peerId, kind);
    });

    setConnected(true);
    return () => {
      svc.disconnect();
      setConnected(false);
      setPeers([]);
    };
  }, [auth.user, auth.token, roomId]);

  const getParticipants = useCallback(async () => {
    if (!signaling) return [];
    try {
      const data = await signaling.getParticipants(roomId);
      setPeers(data.map((d: any) => ({ id: d.id, name: d.name, role: d.role, joinedAt: new Date(d.joinedAt || Date.now()) })));
      return data;
    } catch (e: any) {
      setError(e.message);
      return [];
    }
  }, [signaling, roomId]);

  return { peers, error, connected, getParticipants };
}
