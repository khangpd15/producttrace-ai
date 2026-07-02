import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientKafka } from '@nestjs/microservices';

@Injectable()
export class KafkaProducerService implements OnModuleInit {
  constructor(
    @Inject('KAFKA_PRODUCER')
    private readonly client: ClientKafka,
  ) {}

  async onModuleInit() {
    await this.client.connect();
  }

  emit(topic: string, message: any) {
    return this.client.emit(topic, message);
  }
}