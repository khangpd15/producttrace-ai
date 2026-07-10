import { Controller, Logger } from "@nestjs/common";
import {
  Ctx,
  EventPattern,
  Payload,
  RmqContext,
} from "@nestjs/microservices";

import { SyncService } from "./sync.service";
import { RABBITMQ } from "../../integrations/rabbitmq/rabbitmq.constants";

@Controller()
export class SyncConsumer {
  private readonly logger = new Logger(SyncConsumer.name);

  constructor(
    private readonly syncService: SyncService,
  ) {}

  @EventPattern(RABBITMQ.ROUTING_KEYS.EMBEDDING_GENERATED)
  async consumeEmbeddingGenerated(
    @Payload() payload: unknown,
    @Ctx() context: RmqContext,
  ) {
    const message = context.getMessage();

    const event = this.normalizeEvent(payload);

    if (!event) {
      this.logger.error(
        "Empty event received",
      );
      return;
    }

    this.logger.log(
      `Received rabbitmq event id=${event?.eventId ?? "unknown"} type=${event?.eventType ?? "unknown"}`,
    );

    try {
      await this.syncService.process(event);

      this.logger.log(
        `[SYNC] SUCCESS point=${event?.pointId ?? "unknown"}`,
      );

      // xác nhận xử lý thành công
      const channel = context.getChannelRef();
      channel.ack(message);

    } catch (error) {
      const errMsg =
        error instanceof Error
          ? error.message
          : JSON.stringify(error);

      this.logger.error(
        `[SYNC] FAILED id=${event?.eventId ?? "unknown"} error=${errMsg}`,
      );


      return;
    }
  }

  private normalizeEvent(payload: any): any {
    let data = payload;

    // Buffer -> string
    if (Buffer.isBuffer(data)) {
      data = data.toString("utf8");
    }

    // string -> JSON
    if (typeof data === "string") {
      try {
        data = JSON.parse(data);
      } catch (e) {
        this.logger.error(
          "Invalid JSON payload",
        );
        return null;
      }
    }

    return data;
  }
}