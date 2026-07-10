import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientProxy } from '@nestjs/microservices';

@Injectable()
export class RabbitMQProducerService implements OnModuleInit {
  constructor(
    @Inject('RABBITMQ_PRODUCER')
    private readonly client: ClientProxy,
  ) {}

  async onModuleInit() {
    await this.client.connect();
  }

  emit(pattern: string, message: any) {
    return this.client.emit(pattern, message);
  }
}