"use client";
import { useEffect, useState, useRef } from 'react';
import SignalingService from '@/services/signaling.service';
import { useAuthStore } from '@/store/auth.store';

interface ChatMessage {
  id: string;
  userId: string;
  name?: string;
  text: string;
  timestamp: number;
}

interface ChatPanelProps {
  signaling?: SignalingService | null;
  roomId: string;
  className?: string;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ signaling, roomId, className }) => {
  const auth = useAuthStore();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [connected, setConnected] = useState(false);
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!signaling) return;
    const socket: any = (signaling as any).socket;
    if (!socket) return;

    (async () => {
      try {
        if ((signaling as any).getChatHistory) {
          const history = await (signaling as any).getChatHistory(roomId);
          if (Array.isArray(history)) {
            setMessages(history.map((h: any) => ({
              id: h.id || crypto.randomUUID(),
              userId: h.userId || 'unknown',
              name: h.name,
              text: h.text,
              timestamp: h.timestamp || Date.now(),
            })));
          }
        }
      } catch (err) {
        console.warn('Failed to load chat history', err);
      }
    })();

    const handleChatMessage = (msg: any) => {
      setMessages(prev => [
        ...prev,
        {
          id: msg.id || crypto.randomUUID(),
          userId: msg.userId || 'unknown',
          name: msg.name,
          text: msg.text,
          timestamp: msg.timestamp || Date.now(),
        },
      ]);
    };

    socket.on('chatMessage', handleChatMessage);
    setConnected(true);
    return () => {
      socket.off('chatMessage', handleChatMessage);
    };
  }, [signaling, roomId]);

  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight;
    }
  }, [messages]);

  const sendMessage = () => {
    if (!input.trim() || !signaling) return;
    const socket: any = (signaling as any).socket;
    if (!socket) return;
    const payload: ChatMessage = {
      id: crypto.randomUUID(),
      userId: auth.user?.id || 'unknown',
      name: auth.user?.firstName || 'Me',
      text: input.trim(),
      timestamp: Date.now(),
    };
    socket.emit('chatMessage', { ...payload, roomId });
    setMessages(prev => [...prev, payload]);
    setInput("");
  };

  return (
    <div className={`bg-gray-900 rounded-lg p-4 flex flex-col h-96 ${className || ''}`}>
      <div className="flex items-center gap-2 mb-4">
        <span className="text-blue-400 font-semibold">💬</span>
        <h3 className="text-white font-semibold">Chat {connected ? '' : '(connecting...)'}</h3>
      </div>
      <div ref={listRef} className="flex-1 overflow-y-auto space-y-3 pr-1">
        {messages.map(m => (
          <div key={m.id} className="text-sm">
            <span className="font-medium text-white">{m.name || m.userId.slice(0,6)}</span>:
            <span className="text-gray-300 ml-1 break-words">{m.text}</span>
            <span className="text-[10px] text-gray-500 ml-2">{new Date(m.timestamp).toLocaleTimeString()}</span>
          </div>
        ))}
        {messages.length === 0 && <div className="text-gray-500 text-sm">No messages yet.</div>}
      </div>
      <div className="pt-3 flex items-center gap-2">
        <input
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') sendMessage(); }}
          placeholder="Type a message"
          className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-600"
        />
        <button
          onClick={sendMessage}
          disabled={!input.trim()}
          className="px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 rounded text-white text-sm flex items-center gap-1"
        >
          <span>➤</span> Send
        </button>
      </div>
    </div>
  );
};
