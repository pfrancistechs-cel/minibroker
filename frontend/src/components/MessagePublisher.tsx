'use client';

import React, { useState } from 'react';
import { Send, Sparkles, Key, FileCode, CheckCircle2, Loader2 } from 'lucide-react';
import { publishMessage } from '@/lib/api';

interface MessagePublisherProps {
  topics: string[];
  selectedTopic: string;
  onPublished: () => void;
}

export function MessagePublisher({
  topics,
  selectedTopic,
  onPublished,
}: MessagePublisherProps) {
  const [topic, setTopic] = useState(selectedTopic || (topics[0] ?? ''));
  const [key, setKey] = useState('');
  const [value, setValue] = useState(
    JSON.stringify(
      {
        event_id: `evt_${Math.random().toString(36).substring(2, 9)}`,
        status: 'SUCCESS',
        amount: 299.99,
        timestamp: new Date().toISOString(),
      },
      null,
      2
    )
  );
  const [publishing, setPublishing] = useState(false);
  const [lastResult, setLastResult] = useState<{ partition: number; offset: number } | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);

  // Sync selected topic if changed externally
  React.useEffect(() => {
    if (selectedTopic) setTopic(selectedTopic);
  }, [selectedTopic]);

  const handlePublish = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!topic || !value.trim()) return;

    setPublishing(true);
    setError(null);
    setLastResult(null);

    try {
      const res = await publishMessage(topic, {
        key: key.trim() || undefined,
        value: value.trim(),
        headers: {
          producer: 'minibroker-dashboard',
          client_time: new Date().toISOString(),
        },
      });
      setLastResult({ partition: res.partition, offset: res.offset });
      onPublished();
    } catch (err: any) {
      setError(err?.message || 'Failed to publish');
    } finally {
      setPublishing(false);
    }
  };

  const generateSamplePayload = () => {
    setValue(
      JSON.stringify(
        {
          order_id: `ord_${Math.floor(10000 + Math.random() * 90000)}`,
          customer_id: `cust_${Math.floor(100 + Math.random() * 900)}`,
          currency: 'USD',
          amount: parseFloat((Math.random() * 1500).toFixed(2)),
          items: ['MacBook Pro M3', 'Magic Mouse'],
          status: 'PENDING_PROCESSING',
        },
        null,
        2
      )
    );
    setKey(`cust_${Math.floor(100 + Math.random() * 900)}`);
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Send className="w-4 h-4 text-emerald-400" />
          <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
            Produce Message
          </h2>
        </div>
        <button
          type="button"
          onClick={generateSamplePayload}
          className="flex items-center gap-1 text-[11px] font-mono text-emerald-400 hover:text-emerald-300 bg-emerald-950/40 border border-emerald-800/40 px-2.5 py-1 rounded-lg transition-colors"
        >
          <Sparkles className="w-3 h-3" />
          <span>Random JSON</span>
        </button>
      </div>

      <form onSubmit={handlePublish} className="space-y-3">
        {error && <div className="text-xs text-rose-400 font-mono">{error}</div>}

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {/* Target Topic */}
          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Target Topic
            </label>
            <select
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              className="w-full px-3 py-2 bg-slate-950/80 border border-slate-800 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-emerald-500"
            >
              {topics.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
          </div>

          {/* Partition Key */}
          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Partition Key (Optional)
            </label>
            <div className="relative">
              <input
                type="text"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                placeholder="e.g. user_99 (Hash routed)"
                className="w-full pl-8 pr-3 py-2 bg-slate-950/80 border border-slate-800 rounded-xl text-xs text-white font-mono placeholder:text-slate-600 focus:outline-none focus:border-emerald-500"
              />
              <Key className="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2.5" />
            </div>
          </div>
        </div>

        {/* Message Value Payload */}
        <div>
          <div className="flex items-center justify-between mb-1">
            <label className="text-[11px] font-mono text-slate-400 flex items-center gap-1">
              <FileCode className="w-3 h-3" />
              <span>Payload (JSON or Plain Text)</span>
            </label>
            <span className="text-[10px] text-slate-500 font-mono">
              {key ? 'FNV-1a Consistent Hashing' : 'Round-Robin Partitioning'}
            </span>
          </div>
          <textarea
            required
            rows={5}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            className="w-full px-3 py-2 bg-slate-950/80 border border-slate-800 rounded-xl text-xs text-emerald-400 font-mono leading-relaxed focus:outline-none focus:border-emerald-500"
          />
        </div>

        {/* Submit & Result Bar */}
        <div className="flex items-center justify-between pt-1">
          <div>
            {lastResult && (
              <div className="flex items-center gap-1.5 text-xs font-mono text-emerald-400 animate-in fade-in">
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                <span>
                  Appended to <strong>P-{lastResult.partition}</strong> @ offset{' '}
                  <strong>{lastResult.offset}</strong>
                </span>
              </div>
            )}
          </div>

          <button
            type="submit"
            disabled={publishing || !topic}
            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium rounded-xl shadow-xs transition-colors disabled:opacity-50"
          >
            {publishing ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                <span>Publishing...</span>
              </>
            ) : (
              <>
                <Send className="w-3.5 h-3.5" />
                <span>Publish Message</span>
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
