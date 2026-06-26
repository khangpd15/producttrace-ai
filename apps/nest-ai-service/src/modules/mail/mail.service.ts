import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as sgMail from '@sendgrid/mail';

@Injectable()
export class MailService {
  private readonly logger = new Logger(MailService.name);

  constructor(private configService: ConfigService) {
    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    if (apiKey) {
      sgMail.setApiKey(apiKey);
      this.logger.log('SendGrid API Key is configured');
    } else {
      this.logger.warn('SendGrid API Key is missing in environment variables');
    }
  }

  async sendMail(to: string, subject: string, text: string, html?: string): Promise<boolean> {
    const from = this.configService.get<string>('FROM_EMAIL');
    if (!from) {
      this.logger.error('FROM_EMAIL is missing in environment variables');
      return false;
    }

    const msg = {
      to,
      from,
      subject,
      text,
      html: html || text,
    };

    try {
      await sgMail.send(msg);
      this.logger.log(`Email successfully sent to ${to}`);
      return true;
    } catch (error: any) {
      this.logger.error(`Error sending email to ${to}:`, error);
      if (error.response) {
        this.logger.error(error.response.body);
      }
      return false;
    }
  }
  
  async sendTemplateMail(to: string, templateId: string, dynamicTemplateData?: any): Promise<boolean> {
    const from = this.configService.get<string>('FROM_EMAIL');

    if (!from) {
      this.logger.error('FROM_EMAIL is missing in environment variables');
      return false;
    }

    if (!templateId) {
      this.logger.error('templateId is required but was not provided');
      return false;
    }

    const msg = {
      to,
      from,
      templateId,
      dynamicTemplateData,
    };

    try {
      await sgMail.send(msg);
      this.logger.log(`Template email successfully sent to ${to}`);
      return true;
    } catch (error: any) {
      this.logger.error(`Error sending template email to ${to}:`, error);
      if (error.response) {
        this.logger.error(error.response.body);
      }
      return false;
    }
  }
}
