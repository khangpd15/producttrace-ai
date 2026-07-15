import { Injectable, Logger } from "@nestjs/common";
import { QdrantService } from "../../integrations/qdrant/qdrant.service";
import { RabbitMQProducerService } from "../../messaging/rabbitmq/rabbitmq-producer.service";
import { RABBITMQ } from "../../messaging/rabbitmq/rabbitmq.constants";


@Injectable()
export class SyncService {

    private readonly logger = new Logger(SyncService.name);

    constructor(
        private readonly qdrantService: QdrantService,
        private readonly rabbitmqProducer: RabbitMQProducerService,
    ) {}

    async process(event: any): Promise<void> {

        this.logger.log(`[SYNC] Sync point=${event.pointId}`);

        try {
            this.logger.log(
                `[SYNC][UPSERT] point=${event.pointId} dim=${event.vector?.length ?? 'unknown'}`,
            );

            await this.qdrantService.upsertVector(
                event.pointId,
                event.vector,
                event.payload,
            );

            this.logger.log(
                `[SYNC] Qdrant synced point=${event.pointId}`,
            );
            this.logger.log(
                `[QDRANT][SUCCESS] point=${event.pointId}`,
            );

            await this.rabbitmqProducer.emit(
                RABBITMQ.ROUTING_KEYS.EMBEDDING_COMPLETED,
                {
                    eventId: event.event_id,
                    pointId: event.pointId,
                    timestamp: new Date().toISOString(),
                },
            );

            this.logger.log(
                `[SYNC] embedding.completed published`,
            );

        } catch (error) {

            this.logger.error(
                error instanceof Error
                    ? error.message
                    : JSON.stringify(error),
            );

            throw error;
        }
    }
}