export interface Event<T = unknown> {
  event_id: string;
  event_type: string;
  event_version: string;
  timestamp: string;
  producer: string;
  correlation_id: string;
  payload: T;
}