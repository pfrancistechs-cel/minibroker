'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { TopicDetail, ConsumerGroupSummary, ClusterMetrics } from '@/types';
import { fetchTopics, fetchConsumerGroups, fetchMetrics, subscribeMetricsStream } from '@/lib/api';
import { Header } from '@/components/Header';
import { TopicList } from '@/components/TopicList';
import { MessagePublisher } from '@/components/MessagePublisher';
import { LiveStreamViewer } from '@/components/LiveStreamViewer';
import { ConsumerGroupLag } from '@/components/ConsumerGroupLag';
import { AlertCircle, RefreshCw } from 'lucide-react';

export default function BrokerDashboard() {
  const [topics, setTopics] = useState<TopicDetail[]>([]);
  const [selectedTopic, setSelectedTopic] = useState('orders.created');
  const [consumerGroups, setConsumerGroups] = useState<ConsumerGroupSummary[]>([]);
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      const [fetchedTopics, fetchedGroups, fetchedMetrics] = await Promise.all([
        fetchTopics(),
        fetchConsumerGroups(),
        fetchMetrics(),
      ]);
      setTopics(fetchedTopics);
      setConsumerGroups(fetchedGroups);
      setMetrics(fetchedMetrics);

      if (fetchedTopics.length > 0 && !selectedTopic) {
        setSelectedTopic(fetchedTopics[0].name);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed loading cluster state');
    } finally {
      setLoading(false);
    }
  }, [selectedTopic]);

  useEffect(() => {
    loadData();

    // Subscribe to live cluster telemetry stream (SSE)
    const unsubscribeMetrics = subscribeMetricsStream((updatedMetrics) => {
      setMetrics(updatedMetrics);
    });

    // Periodic refresh of topic watermarks & consumer group lags
    const interval = setInterval(() => {
      loadData();
    }, 4000);

    return () => {
      unsubscribeMetrics();
      clearInterval(interval);
    };
  }, [loadData]);

  const topicNames = topics.map((t) => t.name);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      <Header metrics={metrics} />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {error && (
          <div className="flex items-center justify-between p-4 bg-rose-950/40 border border-rose-800/80 rounded-2xl text-rose-300 text-xs font-mono">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-rose-400" />
              <span>Cluster Error: {error}</span>
            </div>
            <button
              onClick={loadData}
              className="flex items-center gap-1 text-slate-300 hover:text-white px-2 py-1 rounded bg-slate-900 border border-slate-700"
            >
              <RefreshCw className="w-3 h-3" />
              <span>Retry</span>
            </button>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Left Column: Topics and Publisher (5 Cols) */}
          <div className="lg:col-span-5 space-y-6">
            <TopicList
              topics={topics}
              selectedTopic={selectedTopic}
              onSelectTopic={setSelectedTopic}
              onTopicCreated={loadData}
            />

            <MessagePublisher
              topics={topicNames}
              selectedTopic={selectedTopic}
              onPublished={loadData}
            />
          </div>

          {/* Right Column: Live Stream Console & Consumer Lag (7 Cols) */}
          <div className="lg:col-span-7 space-y-6">
            <LiveStreamViewer topic={selectedTopic} />

            <ConsumerGroupLag groups={consumerGroups} />
          </div>
        </div>
      </main>

      <footer className="border-t border-slate-900 py-6 text-center text-xs text-slate-600 font-mono">
        MiniBroker v1.0 • Partitioned Commit Log Architecture with Neon PostgreSQL • MIT License
      </footer>
    </div>
  );
}
