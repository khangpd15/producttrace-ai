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

  emit<T>(pattern: string, message: T) {
    return this.client.emit(pattern, message);
  }
}