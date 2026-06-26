import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { MailService } from './mail.service';
import { SendgridService } from './sendgrid.service';

@Module({
  imports: [ConfigModule],
  providers: [MailService, SendgridService],
  exports: [MailService],
})
export class MailModule {}
