import { Module } from '@nestjs/common';

import { ProductCreatedConsumer } from './messaging/consumers/product-created.consumer';

@Module({
  controllers: [ProductCreatedConsumer],
})
export class AppModule {}