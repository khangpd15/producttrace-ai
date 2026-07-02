import { Injectable } from '@nestjs/common';
import * as sgMail from '@sendgrid/mail';
import { ConfigService } from '@nestjs/config';

@Injectable()
export class SendgridService {
  constructor(private configService: ConfigService) {
    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    if (apiKey) {
      sgMail.setApiKey(apiKey);
    } else {
      console.warn('SendGrid API key is not configured.');
    }
  }

  async sendEmail(to: string, subject: string, html: string): Promise<void> {
    const from = this.configService.get<string>('FROM_EMAIL') || 'noreply@domain.com';
    const msg = {
      to,
      from,
      subject,
      html,
    };

    try {
      if (this.configService.get<string>('SENDGRID_API_KEY')) {
        await sgMail.send(msg);
      } else {
        console.log('Mock email send:', msg);
      }
    } catch (error: any) {
      console.error('Error sending email via SendGrid', error);
      if (error.response) {
        console.error(error.response.body);
      }
    }
  }

  async sendEmailWithTemplate(to: string, templateId: string, dynamicTemplateData: any): Promise<void> {
    const from = this.configService.get<string>('FROM_EMAIL') || 'noreply@domain.com';
    const msg = {
      to,
      from,
      templateId,
      dynamicTemplateData,
    };

    try {
      if (this.configService.get<string>('SENDGRID_API_KEY')) {
        await sgMail.send(msg);
      } else {
        console.log('Mock email send with template:', msg);
      }
    } catch (error: any) {
      console.error('Error sending email via SendGrid Template', error);
      if (error.response) {
        console.error(error.response.body);
      }
    }
  }
}
