import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { EmbeddingModule } from './modules/embedding/embedding.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    EmbeddingModule,
  ],
  controllers: [],
})
export class AppModule {}