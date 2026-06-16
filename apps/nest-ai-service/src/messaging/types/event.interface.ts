export interface Event<T = unknown> {
  eventId: string;
  eventType: string;
  eventVersion: string;
  timestamp: string;
  producer: string;
  correlationId: string;
  payload: T;
}