export interface Message {
  offset: number;
  partition: number;
  topic: string;
  key?: string;
  value: string;
  headers?: Record<string, string>;
  timestamp: string;
}

export interface PartitionInfo {
  id: number;
  high_watermark: number;
  total_messages: number;
}

export interface TopicDetail {
  name: string;
  partitions: number;
  partitions_info: PartitionInfo[];
  total_messages: number;
  created_at: string;
}

export interface ConsumerOffset {
  group_name: string;
  topic: string;
  partition: number;
  offset: number;
  high_watermark: number;
  lag: number;
  updated_at: string;
}

export interface ConsumerGroupSummary {
  group_name: string;
  topic: string;
  offsets: ConsumerOffset[];
  total_lag: number;
}

export interface ClusterMetrics {
  total_messages: number;
  messages_per_sec: number;
  active_topics: number;
  active_groups: number;
  active_streams: number;
  uptime_seconds: number;
}

export interface PublishInput {
  key?: string;
  value: string;
  headers?: Record<string, string>;
}
