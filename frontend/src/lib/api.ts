import { TopicDetail, Message, ConsumerGroupSummary, ClusterMetrics, PublishInput } from '@/types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api';

export async function fetchTopics(): Promise<TopicDetail[]> {
  const res = await fetch(`${API_BASE}/topics`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch topics');
  return res.json();
}

export async function createTopic(name: string, partitions: number): Promise<TopicDetail> {
  const res = await fetch(`${API_BASE}/topics`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, partitions }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create topic');
  }
  return res.json();
}

export async function publishMessage(
  topic: string,
  input: PublishInput
): Promise<{ topic: string; partition: number; offset: number }> {
  const res = await fetch(`${API_BASE}/topics/${encodeURIComponent(topic)}/publish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to publish message');
  }
  return res.json();
}

export async function fetchMessages(
  topic: string,
  partition: number,
  offset = 0,
  limit = 50
): Promise<Message[]> {
  const res = await fetch(
    `${API_BASE}/topics/${encodeURIComponent(topic)}/partitions/${partition}/messages?offset=${offset}&limit=${limit}`,
    { cache: 'no-store' }
  );
  if (!res.ok) throw new Error('Failed to fetch partition messages');
  return res.json();
}

export async function fetchConsumerGroups(): Promise<ConsumerGroupSummary[]> {
  const res = await fetch(`${API_BASE}/consumer-groups`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch consumer groups');
  return res.json();
}

export async function fetchMetrics(): Promise<ClusterMetrics> {
  const res = await fetch(`${API_BASE}/metrics`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch cluster metrics');
  return res.json();
}

export function subscribeTopicStream(
  topic: string,
  onMessage: (msg: Message) => void
): () => void {
  if (typeof window === 'undefined') return () => {};

  const url = `${API_BASE}/topics/${encodeURIComponent(topic)}/stream`;
  const es = new EventSource(url);

  es.addEventListener('message', (event) => {
    try {
      const msg: Message = JSON.parse(event.data);
      onMessage(msg);
    } catch {
      // ignore
    }
  });

  return () => {
    es.close();
  };
}

export function subscribeMetricsStream(
  onMetrics: (m: ClusterMetrics) => void
): () => void {
  if (typeof window === 'undefined') return () => {};

  const url = `${API_BASE}/metrics/stream`;
  const es = new EventSource(url);

  es.onmessage = (event) => {
    try {
      const metrics: ClusterMetrics = JSON.parse(event.data);
      onMetrics(metrics);
    } catch {
      // ignore
    }
  };

  return () => {
    es.close();
  };
}

