'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Message } from '@/types';
import { subscribeTopicStream } from '@/lib/api';
import { Radio, Play, Pause, Trash2, ArrowDown, Hash, Clock, Key } from 'lucide-react';

interface LiveStreamViewerProps {
  topic: string;
}

export function LiveStreamViewer({ topic }: LiveStreamViewerProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Subscribe to Server-Sent Events for active topic
  useEffect(() => {
    if (!topic) return;

    const unsubscribe = subscribeTopicStream(topic, (newMsg) => {
      if (!isPaused) {
        setMessages((prev) => [newMsg, ...prev.slice(0, 199)]);
      }
    });

    return () => {
      unsubscribe();
    };
  }, [topic, isPaused]);

  // Scroll to top when new messages arrive if autoScroll is enabled
  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = 0;
    }
  }, [messages, autoScroll]);

  const partitionColors: Record<number, string> = {
    0: 'bg-cyan-950/80 text-cyan-400 border-cyan-800/60',
    1: 'bg-indigo-950/80 text-indigo-400 border-indigo-800/60',
    2: 'bg-emerald-950/80 text-emerald-400 border-emerald-800/60',
    3: 'bg-amber-950/80 text-amber-400 border-amber-800/60',
  };

  const getPartitionClass = (part: number) => {
    return (
      partitionColors[part % 4] ||
      'bg-slate-800 text-slate-300 border-slate-700'
    );
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4 flex flex-col h-[520px]">
      {/* Stream Controls Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 pb-2 border-b border-slate-800">
        <div className="flex items-center gap-2">
          <span className="relative flex h-2.5 w-2.5">
            <span
              className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                isPaused ? 'bg-amber-400' : 'bg-emerald-400'
              }`}
            ></span>
            <span
              className={`relative inline-flex rounded-full h-2.5 w-2.5 ${
                isPaused ? 'bg-amber-500' : 'bg-emerald-500'
              }`}
            ></span>
          </span>
          <Radio className="w-4 h-4 text-cyan-400" />
          <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
            Live SSE Stream: <span className="font-mono text-cyan-400 normal-case">{topic}</span>
          </h2>
        </div>

        <div className="flex items-center gap-2 text-xs font-mono">
          {/* Pause / Resume */}
          <button
            onClick={() => setIsPaused(!isPaused)}
            className={`flex items-center gap-1 px-2.5 py-1 rounded-lg border transition-colors ${
              isPaused
                ? 'bg-amber-950/40 text-amber-400 border-amber-800/50 hover:bg-amber-900/40'
                : 'bg-slate-800/60 text-slate-300 border-slate-700 hover:bg-slate-800'
            }`}
          >
            {isPaused ? <Play className="w-3 h-3" /> : <Pause className="w-3 h-3" />}
            <span>{isPaused ? 'Resume' : 'Pause'}</span>
          </button>

          {/* Auto Scroll Toggle */}
          <button
            onClick={() => setAutoScroll(!autoScroll)}
            className={`flex items-center gap-1 px-2.5 py-1 rounded-lg border transition-colors ${
              autoScroll
                ? 'bg-cyan-950/40 text-cyan-400 border-cyan-800/50'
                : 'bg-slate-800/60 text-slate-400 border-slate-700'
            }`}
          >
            <ArrowDown className="w-3 h-3" />
            <span>Auto-Scroll</span>
          </button>

          {/* Clear Buffer */}
          <button
            onClick={() => setMessages([])}
            className="p-1 text-slate-400 hover:text-rose-400 rounded-lg hover:bg-rose-950/30 transition-colors"
            title="Clear buffer"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Messages Console */}
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto space-y-2 pr-1 font-mono text-xs"
      >
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-center p-8 text-slate-500 space-y-2">
            <Radio className="w-8 h-8 text-slate-600 animate-pulse" />
            <p className="text-sm">Listening for real-time messages on &quot;{topic}&quot;...</p>
            <p className="text-xs text-slate-600">
              Publish a message on the left to see it stream live via Server-Sent Events.
            </p>
          </div>
        ) : (
          messages.map((m, idx) => (
            <div
              key={`${m.partition}-${m.offset}-${idx}`}
              className="p-3 bg-slate-950/60 border border-slate-800/80 rounded-xl space-y-2 hover:border-slate-700 transition-colors"
            >
              {/* Message Metadata Header */}
              <div className="flex flex-wrap items-center justify-between gap-2 text-[11px] text-slate-400">
                <div className="flex items-center gap-2">
                  <span
                    className={`px-2 py-0.5 rounded border text-[10px] font-bold ${getPartitionClass(
                      m.partition
                    )}`}
                  >
                    P-{m.partition}
                  </span>
                  <span className="flex items-center gap-1 text-slate-300">
                    <Hash className="w-3 h-3 text-slate-500" />
                    <span>offset: {m.offset}</span>
                  </span>
                  {m.key && (
                    <span className="flex items-center gap-1 text-amber-400">
                      <Key className="w-3 h-3 text-amber-500" />
                      <span>key: {m.key}</span>
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-1 text-slate-500 text-[10px]">
                  <Clock className="w-3 h-3" />
                  <span>{new Date(m.timestamp).toLocaleTimeString()}</span>
                </div>
              </div>

              {/* Payload Body */}
              <pre className="text-xs text-slate-200 bg-slate-900/80 p-2.5 rounded-lg overflow-x-auto whitespace-pre-wrap leading-relaxed border border-slate-800/60">
                {m.value}
              </pre>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

