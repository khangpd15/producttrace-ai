import { Controller, Post, Body, HttpStatus, HttpException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { MailService } from './modules/mail/mail.service';

@Controller('mock')
export class MockController {
  constructor(
    private readonly mailService: MailService,
    private readonly configService: ConfigService,
  ) { }

  @Post('register')
  async simulateRegistration(@Body() body: { email: string; name: string }) {
    if (!body.email || !body.name) {
      throw new HttpException('Vui lòng cung cấp email và name trong body JSON', HttpStatus.BAD_REQUEST);
    }

    const templateId = this.configService.get<string>('WELCOME_TEMPLATE_ID');

    if (!templateId) {
      throw new HttpException('Chưa cấu hình WELCOME_TEMPLATE_ID trong .env', HttpStatus.INTERNAL_SERVER_ERROR);
    }

    const success = await this.mailService.sendTemplateMail(body.email, templateId, {
      name: body.name,
    });

    if (success) {
      return {
        message: 'Đã mô phỏng đăng ký và gửi email chào mừng thành công',
        email: body.email,
      };
    } else {
      throw new HttpException('Đăng ký thành công nhưng gửi email thất bại', HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }
}
