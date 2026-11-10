'use client';

import React from 'react';
import { ClusterMetrics } from '@/types';
import { Layers, Activity, Radio, Database, Clock, Server } from 'lucide-react';

interface HeaderProps {
  metrics: ClusterMetrics | null;
}

export function Header({ metrics }: HeaderProps) {
  const formatUptime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}m ${secs}s`;
  };

  return (
    <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-30">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col md:flex-row md:items-center justify-between py-4 gap-4">
          {/* Logo & Cluster Tag */}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center text-white shadow-md shadow-cyan-500/10">
              <Layers className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-lg font-bold text-white tracking-tight">MiniBroker</h1>
                <span className="text-[10px] font-mono uppercase font-bold tracking-widest px-2 py-0.5 rounded-md bg-cyan-950/80 text-cyan-400 border border-cyan-800/60">
                  Kafka Style
                </span>
              </div>
              <p className="text-xs text-slate-400">
                High-Throughput Partitioned Commit Log Engine + Neon PostgreSQL
              </p>
            </div>
          </div>

          {/* Telemetry Stats Bar */}
          <div className="flex flex-wrap items-center gap-2 sm:gap-4 text-xs font-mono">
            {/* Throughput */}
            <div className="flex items-center gap-2 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              <Activity className="w-3.5 h-3.5 text-emerald-400" />
              <span className="text-slate-400">Rate:</span>
              <strong className="text-emerald-400">
                {metrics ? metrics.messages_per_sec.toFixed(1) : '0.0'} msg/s
              </strong>
            </div>

            {/* Total Messages */}
            <div className="flex items-center gap-1.5 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl">
              <Database className="w-3.5 h-3.5 text-blue-400" />
              <span className="text-slate-400">Total:</span>
              <strong className="text-white">
                {metrics ? metrics.total_messages.toLocaleString() : '0'}
              </strong>
            </div>

            {/* Active Topics */}
            <div className="flex items-center gap-1.5 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl">
              <Server className="w-3.5 h-3.5 text-cyan-400" />
              <span className="text-slate-400">Topics:</span>
              <strong className="text-white">{metrics ? metrics.active_topics : 0}</strong>
            </div>

            {/* SSE Streams */}
            <div className="flex items-center gap-1.5 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl">
              <Radio className="w-3.5 h-3.5 text-amber-400" />
              <span className="text-slate-400">SSE Streams:</span>
              <strong className="text-white">{metrics ? metrics.active_streams : 0}</strong>
            </div>

            {/* Uptime */}
            <div className="hidden lg:flex items-center gap-1.5 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl text-slate-400">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              <span>{metrics ? formatUptime(metrics.uptime_seconds) : '0m'}</span>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

