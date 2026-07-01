import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { SendgridService } from './sendgrid.service';

@Injectable()
export class MailService {
  constructor(
    private readonly sendgridService: SendgridService,
    private readonly configService: ConfigService,
  ) {}

  async sendPasswordResetEmail(to: string, resetUrl: string, name: string): Promise<void> {
    const templateId = this.configService.get<string>('RESET_PASSWORD_TEMPLATE_ID') || 'd-76dced8ef3f6486aa296d2b25899b24e';
    
    // Map variables to match the SendGrid template
    const dynamicTemplateData = {
      name: name,
      resetLink: resetUrl,
      year: new Date().getFullYear().toString(),
    };

    await this.sendgridService.sendEmailWithTemplate(
      to,
      templateId,
      dynamicTemplateData
    );
  }
}
