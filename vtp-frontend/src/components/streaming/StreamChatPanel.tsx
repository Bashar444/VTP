"use client";
import { useEffect, useState, useRef, useCallback } from 'react';
import { useAuthStore } from '@/store/auth.store';

interface ChatMessage {
  id: string;
  session_id: string;
  user_id: string;
  user_name?: string;
  message_type: 'text' | 'question' | 'announcement' | 'system';
  content: string;
  is_pinned: boolean;
  is_answered: boolean;
  reply_to_id?: string;
  created_at: string;
}

interface StreamChatPanelProps {
  sessionId: string;
  roomId: string;
  isInstructor?: boolean;
  className?: string;
}

export const StreamChatPanel: React.FC<StreamChatPanelProps> = ({
  sessionId,
  roomId,
  isInstructor = false,
  className
}) => {
  const auth = useAuthStore();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [messageType, setMessageType] = useState<'text' | 'question'>('text');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [showQuestions, setShowQuestions] = useState(false);
  const listRef = useRef<HTMLDivElement>(null);

  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const fetchMessages = useCallback(async () => {
    try {
      const endpoint = showQuestions
        ? `${API_URL}/api/v1/streaming/sessions/${sessionId}/chat/questions`
        : `${API_URL}/api/v1/streaming/sessions/${sessionId}/chat?recent=true&limit=100`;

      const res = await fetch(endpoint, {
        headers: {
          Authorization: `Bearer ${auth.token}`
        }
      });

      if (res.ok) {
        const data = await res.json();
        const msgs = showQuestions ? data.questions : data.messages;
        setMessages(msgs || []);
      }
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setLoading(false);
    }
  }, [sessionId, showQuestions, API_URL, auth.token]);

  useEffect(() => {
    fetchMessages();
    // Poll for new messages every 3 seconds
    const interval = setInterval(fetchMessages, 3000);
    return () => clearInterval(interval);
  }, [fetchMessages]);

  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight;
    }
  }, [messages]);

  const sendMessage = async () => {
    if (!input.trim() || sending) return;

    try {
      setSending(true);
      const res = await fetch(`${API_URL}/api/v1/streaming/sessions/${sessionId}/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`
        },
        body: JSON.stringify({
          content: input.trim(),
          message_type: messageType
        })
      });

      if (res.ok) {
        const newMsg = await res.json();
        setMessages(prev => [...prev, newMsg]);
        setInput('');
        setMessageType('text');
      } else {
        const error = await res.json();
        console.error('Failed to send message:', error);
      }
    } catch (err) {
      console.error('Failed to send message:', err);
    } finally {
      setSending(false);
    }
  };

  const pinMessage = async (messageId: string, pinned: boolean) => {
    try {
      await fetch(`${API_URL}/api/v1/streaming/sessions/${sessionId}/chat/${messageId}/pin`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`
        },
        body: JSON.stringify({ pinned })
      });
      fetchMessages();
    } catch (err) {
      console.error('Failed to pin message:', err);
    }
  };

  const markAsAnswered = async (messageId: string) => {
    try {
      await fetch(`${API_URL}/api/v1/streaming/sessions/${sessionId}/chat/${messageId}/answer`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${auth.token}`
        }
      });
      fetchMessages();
    } catch (err) {
      console.error('Failed to mark as answered:', err);
    }
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleTimeString('ar-SA', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getMessageBgColor = (msg: ChatMessage) => {
    if (msg.is_pinned) return 'bg-yellow-900/30 border-r-4 border-yellow-500';
    if (msg.message_type === 'announcement') return 'bg-blue-900/30 border-r-4 border-blue-500';
    if (msg.message_type === 'question') {
      return msg.is_answered
        ? 'bg-green-900/20 border-r-4 border-green-500'
        : 'bg-purple-900/20 border-r-4 border-purple-500';
    }
    return 'bg-gray-800/50';
  };

  return (
    <div dir="rtl" className={`bg-gray-900 rounded-lg flex flex-col h-full ${className || ''}`}>
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h3 className="text-white font-semibold flex items-center gap-2">
            💬 المحادثة
          </h3>
          <div className="flex gap-2">
            <button
              onClick={() => setShowQuestions(false)}
              className={`px-3 py-1 text-xs rounded-full ${
                !showQuestions
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              الكل
            </button>
            <button
              onClick={() => setShowQuestions(true)}
              className={`px-3 py-1 text-xs rounded-full ${
                showQuestions
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              الأسئلة
            </button>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div ref={listRef} className="flex-1 overflow-y-auto p-4 space-y-3">
        {loading ? (
          <div className="text-center text-gray-500 py-8">
            <div className="animate-spin w-6 h-6 border-2 border-gray-500 border-t-blue-500 rounded-full mx-auto mb-2"></div>
            جاري تحميل الرسائل...
          </div>
        ) : messages.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            {showQuestions ? 'لا توجد أسئلة بعد' : 'لا توجد رسائل بعد. كن أول من يرسل!'}
          </div>
        ) : (
          messages.map(msg => (
            <div
              key={msg.id}
              className={`rounded-lg p-3 ${getMessageBgColor(msg)}`}
            >
              {/* Message header */}
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  {msg.message_type === 'announcement' && (
                    <span className="text-blue-400 text-xs">📢 إعلان</span>
                  )}
                  {msg.message_type === 'question' && (
                    <span className="text-purple-400 text-xs">
                      {msg.is_answered ? '✓ سؤال مُجاب' : '❓ سؤال'}
                    </span>
                  )}
                  {msg.is_pinned && (
                    <span className="text-yellow-400 text-xs">📌 مثبت</span>
                  )}
                  <span className="text-white font-medium text-sm">
                    {msg.user_name || 'مستخدم'}
                  </span>
                </div>
                <span className="text-gray-500 text-xs">{formatTime(msg.created_at)}</span>
              </div>

              {/* Message content */}
              <p className="text-gray-200 text-sm break-words">{msg.content}</p>

              {/* Instructor actions */}
              {isInstructor && (
                <div className="flex gap-2 mt-2">
                  <button
                    onClick={() => pinMessage(msg.id, !msg.is_pinned)}
                    className="text-xs text-gray-400 hover:text-yellow-400"
                  >
                    {msg.is_pinned ? '📌 إلغاء التثبيت' : '📌 تثبيت'}
                  </button>
                  {msg.message_type === 'question' && !msg.is_answered && (
                    <button
                      onClick={() => markAsAnswered(msg.id)}
                      className="text-xs text-gray-400 hover:text-green-400"
                    >
                      ✓ تم الرد
                    </button>
                  )}
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* Input */}
      <div className="p-4 border-t border-gray-700">
        {/* Message type selector */}
        <div className="flex gap-2 mb-2">
          <button
            onClick={() => setMessageType('text')}
            className={`px-3 py-1 text-xs rounded ${
              messageType === 'text'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300'
            }`}
          >
            💬 رسالة
          </button>
          <button
            onClick={() => setMessageType('question')}
            className={`px-3 py-1 text-xs rounded ${
              messageType === 'question'
                ? 'bg-purple-600 text-white'
                : 'bg-gray-700 text-gray-300'
            }`}
          >
            ❓ سؤال
          </button>
        </div>

        <div className="flex gap-2">
          <input
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) sendMessage(); }}
            placeholder={messageType === 'question' ? 'اكتب سؤالك...' : 'اكتب رسالتك...'}
            className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-600 placeholder-gray-500"
          />
          <button
            onClick={sendMessage}
            disabled={!input.trim() || sending}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-white text-sm font-medium"
          >
            {sending ? '...' : 'إرسال'}
          </button>
        </div>
      </div>
    </div>
  );
};
