'use client';

import React, { useState } from 'react';
import { TopicDetail } from '@/types';
import { Server, Plus, Database, ChevronRight, Hash, Loader2 } from 'lucide-react';
import { createTopic } from '@/lib/api';

interface TopicListProps {
  topics: TopicDetail[];
  selectedTopic: string;
  onSelectTopic: (topic: string) => void;
  onTopicCreated: () => void;
}

export function TopicList({
  topics,
  selectedTopic,
  onSelectTopic,
  onTopicCreated,
}: TopicListProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [newTopicName, setNewTopicName] = useState('');
  const [partitions, setPartitions] = useState('3');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTopicName.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      await createTopic(newTopicName.trim(), parseInt(partitions, 10) || 3);
      onTopicCreated();
      onSelectTopic(newTopicName.trim());
      setIsCreating(false);
      setNewTopicName('');
    } catch (err: any) {
      setError(err?.message || 'Failed to create topic');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Server className="w-4 h-4 text-cyan-400" />
          <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
            Topics ({topics.length})
          </h2>
        </div>
        <button
          onClick={() => setIsCreating(!isCreating)}
          className="flex items-center gap-1 text-xs text-cyan-400 hover:text-cyan-300 font-medium px-2 py-1 rounded-lg bg-cyan-950/40 border border-cyan-800/40 transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>New Topic</span>
        </button>
      </div>

      {/* Create Topic Form Drawer */}
      {isCreating && (
        <form
          onSubmit={handleCreate}
          className="p-3.5 bg-slate-950/80 border border-slate-800 rounded-xl space-y-3 animate-in fade-in duration-150"
        >
          <div className="text-xs font-semibold text-slate-300">Create New Topic</div>
          {error && <div className="text-xs text-rose-400 font-mono">{error}</div>}

          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Topic Name
            </label>
            <input
              type="text"
              required
              value={newTopicName}
              onChange={(e) => setNewTopicName(e.target.value)}
              placeholder="e.g. telemetry.events"
              className="w-full px-3 py-1.5 bg-slate-900 border border-slate-800 rounded-lg text-xs text-white placeholder:text-slate-500 font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div>
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Partitions (1-16)
            </label>
            <input
              type="number"
              min="1"
              max="16"
              value={partitions}
              onChange={(e) => setPartitions(e.target.value)}
              className="w-full px-3 py-1.5 bg-slate-900 border border-slate-800 rounded-lg text-xs text-white font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={() => setIsCreating(false)}
              className="px-2.5 py-1 text-xs text-slate-400 hover:text-slate-200"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex items-center gap-1.5 px-3 py-1 bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-medium rounded-lg transition-colors disabled:opacity-50"
            >
              {submitting && <Loader2 className="w-3 h-3 animate-spin" />}
              <span>Provision</span>
            </button>
          </div>
        </form>
      )}

      {/* Topics List */}
      <div className="space-y-1.5 max-h-[380px] overflow-y-auto pr-1">
        {topics.map((t) => {
          const isSelected = selectedTopic === t.name;
          return (
            <button
              key={t.name}
              onClick={() => onSelectTopic(t.name)}
              className={`w-full text-left p-3 rounded-xl border transition-all flex items-center justify-between group ${
                isSelected
                  ? 'bg-cyan-950/30 border-cyan-500/50 text-white shadow-xs'
                  : 'bg-slate-950/40 border-slate-800/80 text-slate-300 hover:border-slate-700 hover:bg-slate-950/80'
              }`}
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs font-semibold group-hover:text-cyan-400 transition-colors">
                    {t.name}
                  </span>
                </div>
                <div className="flex items-center gap-3 text-[11px] font-mono text-slate-400">
                  <span className="flex items-center gap-1">
                    <Hash className="w-3 h-3 text-cyan-400" />
                    <span>{t.partitions} parts</span>
                  </span>
                  <span className="flex items-center gap-1">
                    <Database className="w-3 h-3 text-blue-400" />
                    <span>{t.total_messages.toLocaleString()} msgs</span>
                  </span>
                </div>
              </div>

              <ChevronRight
                className={`w-4 h-4 transition-transform ${
                  isSelected ? 'text-cyan-400 translate-x-0.5' : 'text-slate-600'
                }`}
              />
            </button>
          );
        })}
      </div>
    </div>
  );
}
