import { Module } from '@nestjs/common';
import { AuthService } from './auth.service';
import { AuthController } from './auth.controller';
import { USER_REPOSITORY } from './repositories/user.repository';
import { JsonUserRepository } from './repositories/json-user.repository';
import { PASSWORD_RESET_REPOSITORY } from './repositories/password-reset.repository';
import { JsonPasswordResetRepository } from './repositories/json-password-reset.repository';
import { MailModule } from '../mail/mail.module';

@Module({
  imports: [MailModule],
  controllers: [AuthController],
  providers: [
    AuthService,
    {
      provide: USER_REPOSITORY,
      useClass: JsonUserRepository,
    },
    {
      provide: PASSWORD_RESET_REPOSITORY,
      useClass: JsonPasswordResetRepository,
    },
  ],
})
export class AuthModule {}
