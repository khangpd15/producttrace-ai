import { Controller, Logger } from "@nestjs/common";
import {
  Ctx,
  EventPattern,
  Payload,
  KafkaContext,
} from "@nestjs/microservices";

import { SyncService } from "./sync.service";
import { KAFKA } from "../../kafka/kafka.constants";

@Controller()
export class SyncConsumer {
  private readonly logger = new Logger(SyncConsumer.name);

  constructor(private readonly syncService: SyncService) {}

  @EventPattern(KAFKA.TOPICS.EMBEDDING_GENERATED)
  async consumeEmbeddingGenerated(
    @Payload() payload: unknown,
    @Ctx() context: KafkaContext,
  ) {
    const topic = context.getTopic();
    const partition = context.getPartition();
    const message = context.getMessage();
    const offset = message?.offset ?? "unknown";

    const event = this.normalizeEvent(payload);

    if (!event) {
      this.logger.error(
        `Empty event received topic=${topic} partition=${partition} offset=${offset}`,
      );
      return;
    }

    this.logger.log(
      `Received kafka event topic=${topic} partition=${partition} offset=${offset}`,
    );

    try {
      await this.syncService.process(event);

      this.logger.log(
        `[SYNC] SUCCESS point=${event?.pointId ?? "unknown"}`,
      );
    } catch (error) {
      const errMsg =
        error instanceof Error ? error.message : JSON.stringify(error);

      this.logger.error(
        `[SYNC] FAILED topic=${topic} partition=${partition} offset=${offset} error=${errMsg}`,
      );

      /**
       * IMPORTANT FIX:
       * ❌ KHÔNG throw lại error -> tránh KafkaJS crash loop
       * ✔ log + swallow hoặc send to DLQ (nếu sau này bạn có)
       */
      return;
    }
  }

  private normalizeEvent(payload: any): any {
    let data = payload;

    // Kafka wrapper
    if (data?.value) {
      data = data.value;
    }

    // Buffer -> string
    if (Buffer.isBuffer(data)) {
      data = data.toString("utf8");
    }

    // string -> JSON
    if (typeof data === "string") {
      try {
        data = JSON.parse(data);
      } catch (e) {
        this.logger.error("Invalid JSON payload");
        return null;
      }
    }

    return data;
  }
}