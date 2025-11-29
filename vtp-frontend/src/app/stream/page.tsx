"use client";
import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function StreamJoinPage() {
  const router = useRouter();
  const [roomId, setRoomId] = useState("");

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    if (!roomId.trim()) return;
    router.push(`/stream/${roomId.trim()}`);
  };

  return (
    <main className="max-w-xl mx-auto pt-32 px-4">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Join a Stream</h1>
      <p className="text-gray-600 mb-8">Enter the room ID provided to you to join a live session.</p>
      <form onSubmit={handleJoin} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1" htmlFor="room">Room ID</label>
          <input
            id="room"
            type="text"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
            placeholder="e.g. math-class-2025"
            className="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 bg-white"
          />
        </div>
        <button
          type="submit"
          disabled={!roomId.trim()}
          className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium py-2 rounded-md transition"
        >
          Join Stream
        </button>
      </form>
      <div className="mt-8 text-sm text-gray-500">
        Need a room? Ask your instructor or create one from the dashboard (coming soon).
      </div>
    </main>
  );
}
