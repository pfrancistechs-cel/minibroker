'use client';

import React from 'react';
import { ConsumerGroupSummary } from '@/types';
import { Users, AlertTriangle, CheckCircle, Hash } from 'lucide-react';

interface ConsumerGroupLagProps {
  groups: ConsumerGroupSummary[];
}

export function ConsumerGroupLag({ groups }: ConsumerGroupLagProps) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="w-4 h-4 text-purple-400" />
          <h2 className="text-sm font-semibold text-white uppercase tracking-wider">
            Consumer Groups & Lag Monitor ({groups.length})
          </h2>
        </div>
        <span className="text-[11px] font-mono text-slate-500">
          Committed in Neon PostgreSQL
        </span>
      </div>

      {groups.length === 0 ? (
        <div className="p-6 text-center text-xs text-slate-500 font-mono">
          No consumer groups registered yet.
        </div>
      ) : (
        <div className="space-y-3">
          {groups.map((g) => {
            return (
              <div
                key={`${g.group_name}-${g.topic}`}
                className="p-4 bg-slate-950/60 border border-slate-800 rounded-xl space-y-3"
              >
                {/* Header row */}
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="space-y-0.5">
                    <div className="font-mono text-xs font-semibold text-white flex items-center gap-2">
                      <span>{g.group_name}</span>
                      <span className="text-[10px] text-slate-400 font-normal">
                        subscribed to <strong className="text-cyan-400">{g.topic}</strong>
                      </span>
                    </div>
                  </div>

                  {/* Total Lag Badge */}
                  <div className="flex items-center gap-1.5 font-mono text-xs">
                    {g.total_lag === 0 ? (
                      <span className="flex items-center gap-1 text-emerald-400 bg-emerald-950/40 border border-emerald-800/40 px-2 py-0.5 rounded-full text-[11px]">
                        <CheckCircle className="w-3 h-3" />
                        <span>Lag: 0 (Caught Up)</span>
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-amber-400 bg-amber-950/40 border border-amber-800/40 px-2 py-0.5 rounded-full text-[11px]">
                        <AlertTriangle className="w-3 h-3" />
                        <span>Total Lag: {g.total_lag}</span>
                      </span>
                    )}
                  </div>
                </div>

                {/* Partition Offset Details */}
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2 pt-1 font-mono text-xs">
                  {g.offsets.map((o) => {
                    const percent =
                      o.high_watermark === 0
                        ? 100
                        : Math.min(100, Math.round((o.offset / o.high_watermark) * 100));

                    return (
                      <div
                        key={o.partition}
                        className="p-2.5 bg-slate-900/80 border border-slate-800/80 rounded-lg space-y-1.5"
                      >
                        <div className="flex items-center justify-between text-[11px] text-slate-400">
                          <span className="flex items-center gap-1 text-cyan-400 font-semibold">
                            <Hash className="w-3 h-3" />
                            <span>P-{o.partition}</span>
                          </span>
                          <span>
                            {o.offset} / {o.high_watermark}
                          </span>
                        </div>

                        {/* Progress Bar */}
                        <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                          <div
                            className={`h-full rounded-full transition-all duration-300 ${
                              o.lag === 0 ? 'bg-emerald-500' : 'bg-amber-500'
                            }`}
                            style={{ width: `${percent}%` }}
                          />
                        </div>

                        <div className="flex justify-between text-[10px] text-slate-500">
                          <span>Progress: {percent}%</span>
                          <span className={o.lag > 0 ? 'text-amber-400' : 'text-slate-500'}>
                            lag: {o.lag}
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
