import { Injectable } from '@nestjs/common';
import * as sgMail from '@sendgrid/mail';
import { ConfigService } from '@nestjs/config';

@Injectable()
export class SendgridService {
  constructor(private configService: ConfigService) {
    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    if (apiKey && !apiKey.includes('your_')) {
      sgMail.setApiKey(apiKey);
    } else {
      console.warn('SendGrid API key is not configured or using placeholder. Running in MOCK mode.');
    }
  }

  async sendEmail(to: string, subject: string, html: string): Promise<void> {
    const from = this.configService.get<string>('SENDGRID_FROM_EMAIL') || 
                 this.configService.get<string>('FROM_EMAIL') || 
                 'noreply@producttrace-ai.com';
    const msg = {
      to,
      from,
      subject,
      html,
    };

    try {
      const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
      if (apiKey && !apiKey.includes('your_')) {
        await sgMail.send(msg);
      } else {
        console.log('[MOCK EMAIL SENT]', msg);
      }
    } catch (error: any) {
      console.error('Error sending email via SendGrid', error);
      if (error.response) {
        console.error(error.response.body);
      }
    }
  }

  async sendEmailWithTemplate(to: string, templateId: string, dynamicTemplateData: any): Promise<void> {
    const from = this.configService.get<string>('SENDGRID_FROM_EMAIL') || 
                 this.configService.get<string>('FROM_EMAIL') || 
                 'noreply@producttrace-ai.com';
    const msg = {
      to,
      from,
      templateId,
      dynamicTemplateData,
    };

    try {
      const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
      if (apiKey && !apiKey.includes('your_') && templateId && !templateId.includes('your_') && !templateId.includes('your-')) {
        await sgMail.send(msg);
      } else {
        console.log('[MOCK TEMPLATE EMAIL SENT]', msg);
      }
    } catch (error: any) {
      console.error('Error sending email via SendGrid Template', error);
      if (error.response) {
        console.error(error.response.body);
      }
    }
  }
}
