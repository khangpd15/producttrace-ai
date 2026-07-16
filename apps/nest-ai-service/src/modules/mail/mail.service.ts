import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as sgMail from '@sendgrid/mail';

@Injectable()
export class MailService {
  private readonly logger = new Logger(MailService.name);

  constructor(private configService: ConfigService) {
    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    if (apiKey && !apiKey.includes('your_')) {
      sgMail.setApiKey(apiKey);
      this.logger.log('SendGrid API Key is configured');
    } else {
      this.logger.warn('SendGrid API Key is missing or using placeholder in environment variables. Email will run in MOCK mode.');
    }
  }

  async sendMail(to: string, subject: string, text: string, html?: string): Promise<boolean> {
    const from = this.configService.get<string>('FROM_EMAIL') ||
      this.configService.get<string>('SENDGRID_FROM_EMAIL') ||
      'noreply@producttrace-ai.com';

    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    const isMock = !apiKey || apiKey.includes('your_') || from.includes('example.com');

    if (isMock) {
      this.logger.log(`[MOCK EMAIL SENT]`);
      this.logger.log(`To: ${to}`);
      this.logger.log(`From: ${from}`);
      this.logger.log(`Subject: ${subject}`);
      this.logger.log(`Text Body: ${text}`);
      this.logger.log(`HTML Body: ${html}`);
      return true;
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
      this.logger.log(`Email successfully sent via SendGrid to ${to}`);
      return true;
    } catch (error: any) {
      this.logger.error(`Error sending email to ${to}:`, error);
      if (error.response) {
        this.logger.error(error.response.body);
      }
      return false;
    }
  }

  async sendEmailWithTemplate(to: string, templateId: string, dynamicTemplateData?: any): Promise<boolean> {
    const from = this.configService.get<string>('SENDGRID_FROM_EMAIL') ||
      this.configService.get<string>('FROM_EMAIL') ||
      'noreply@producttrace-ai.com';

    if (!templateId) {
      this.logger.error('templateId is required but was not provided');
      return false;
    }

    const apiKey = this.configService.get<string>('SENDGRID_API_KEY');
    const isMock = !apiKey || apiKey.includes('your_') || from.includes('example.com') || templateId.includes('your_') || templateId.includes('your-');

    if (isMock) {
      this.logger.log(`[MOCK TEMPLATE EMAIL SENT]`);
      this.logger.log(`To: ${to}`);
      this.logger.log(`From: ${from}`);
      this.logger.log(`Template ID: ${templateId}`);
      this.logger.log(`Template Data: ${JSON.stringify(dynamicTemplateData)}`);
      return true;
    }

    const msg = {
      to,
      from,
      templateId,
      dynamicTemplateData,
    };

    try {
      await sgMail.send(msg);
      this.logger.log(`Template email successfully sent via SendGrid to ${to}`);
      return true;
    } catch (error: any) {
      this.logger.error(`Error sending template email to ${to}:`, error);
      if (error.response) {
        this.logger.error(error.response.body);
      }
      return false;
    }
  }

  async sendTemplateMail(to: string, templateId: string, dynamicTemplateData?: any): Promise<boolean> {
    return this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
  }

  async sendVerificationOtpEmail(to: string, fullName: string, otpCode: string): Promise<boolean> {
    const templateId = this.configService.get<string>('VERIFY_EMAIL_TEMPLATE_ID') ||
      this.configService.get<string>('WELCOME_TEMPLATE_ID') ||
      '';

    const dynamicTemplateData = {
      fullName,
      otpCode,
      year: new Date().getFullYear().toString(),
    };

    return this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
  }
  async sendOTPOwnerShipEmail(to: string, productId: string, otpCode: string): Promise<boolean> {
    const templateId = this.configService.get<string>('OWNERSHIP_REGISTER') || '';
    const dynamicTemplateData = {
      productId,
      otpCode,
      year: new Date().getFullYear().toString(),
    };

    return this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
  }




  async sendWelcomeEmail(to: string, fullName: string): Promise<boolean> {
    const templateId = this.configService.get<string>('WELCOME_EMAIL_TEMPLATE_ID') ||
      this.configService.get<string>('WELCOME_TEMPLATE_ID') ||
      '';
    const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
    const loginUrl = `${frontendUrl}/login`;

    const dynamicTemplateData = {
      fullName,
      loginUrl,
      year: new Date().getFullYear().toString(),
    };

    return this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
  }


  async sendOTPOwnerShip(to: string, productId: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('OWNERSHIP_REGISTER');
    if (templateId && !templateId.includes('your_')) {
      await this.sendOTPOwnerShipEmail(to, productId, otpCode);
    } else {
      // Fallback plain email if template is not available
      const subject = 'You have been assigned a product: ' + productId;
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #333;">Welcome to ProductTrace AI!</h2>
          <p>Thank you for registering, <strong>${productId}</strong>.</p>
          <p>Please use the following One-Time Password (OTP) to verify your account. This code is valid for 5 minutes:</p>
          <div style="background-color: #f5f5f5; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; color: #007bff; border-radius: 4px; margin: 20px 0;">
            ${otpCode}
          </div>
          <p>If you did not request this code, please ignore this email.</p>
        </div>
      `;
      await this.sendMail(to, subject, html);
    }
  }


  async sendOTP(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('VERIFY_EMAIL_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_')) {
      await this.sendVerificationOtpEmail(to, fullName, otpCode);
    } else {
      const subject = 'Welcome to ProductTrace AI - Verify Your Account';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #333;">Welcome to ProductTrace AI!</h2>
          <p>Thank you for registering, <strong>${fullName}</strong>.</p>
          <p>Please use the following One-Time Password (OTP) to verify your account. This code is valid for 5 minutes:</p>
          <div style="background-color: #f5f5f5; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; color: #007bff; border-radius: 4px; margin: 20px 0;">
            ${otpCode}
          </div>
          <p>If you did not request this code, please ignore this email.</p>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }

  async sendVerificationSuccess(to: string, fullName: string): Promise<void> {
    const templateId = this.configService.get<string>('WELCOME_EMAIL_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_')) {
      await this.sendWelcomeEmail(to, fullName);
    } else {
      const subject = 'Xác thực tài khoản thành công';
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      const loginUrl = `${frontendUrl}/login`;
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #2e7d32;">Xác thực tài khoản thành công!</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Tài khoản của bạn đã được xác thực thành công tại ProductTrace AI.</p>
          <p>Bây giờ bạn đã có thể bắt đầu sử dụng hệ thống.</p>
          <div style="margin: 20px 0;">
            <a href="${loginUrl}" style="background-color: #2e7d32; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; font-weight: bold;">Đăng nhập ngay</a>
          </div>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }

  async sendPasswordReset(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('RESET_PASSWORD_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_') && !templateId.includes('your-')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      const resetUrl = `${frontendUrl}/reset-password?email=${encodeURIComponent(to)}&code=${otpCode}`;

      const dynamicTemplateData = {
        fullName,
        name: fullName,         // alias for templates using {{name}}
        otpCode,
        resetLink: resetUrl,
        year: new Date().getFullYear().toString(),
      };
      await this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
    } else {
      const subject = 'Reset Your Password - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #333;">Yêu cầu đặt lại mật khẩu</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Chúng tôi nhận được yêu cầu đặt lại mật khẩu cho tài khoản của bạn.</p>
          <p>Vui lòng sử dụng mã OTP bên dưới để hoàn tất việc đặt lại mật khẩu (có hiệu lực trong 5 phút):</p>
          <div style="background-color: #f5f5f5; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; color: #d32f2f; border-radius: 4px; margin: 20px 0;">
            ${otpCode}
          </div>
          <p>Nếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này.</p>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }

  async sendPasswordResetEmail(to: string, resetUrl: string, name: string): Promise<void> {
    const templateId = this.configService.get<string>('RESET_PASSWORD_TEMPLATE_ID') || 'd-76dced8ef3f6486aa296d2b25899b24e';
    const dynamicTemplateData = {
      name: name,
      resetLink: resetUrl,
      year: new Date().getFullYear().toString(),
    };
    await this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
  }

  async sendWarrantyUpdateEmail(to: string, fullName: string, productName: string, status: string, endDate: string): Promise<void> {
    const templateId = this.configService.get<string>('WARRANTY_UPDATE_TEMPLATE_ID') || 'd-aa9b56ba4bf64b54a72eddc7ba33ba03';
    if (templateId && !templateId.includes('your_') && !templateId.includes('your-')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      
      const dynamicTemplateData = {
        fullName,
        productName,
        status,
        endDate,
        frontendUrl,
        year: new Date().getFullYear().toString(),
      };
      await this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
    } else {
      const subject = 'Cập nhật bảo hành sản phẩm - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #1976d2;">Thông báo cập nhật bảo hành</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Trạng thái bảo hành cho sản phẩm của bạn vừa được cập nhật trên hệ thống ProductTrace AI.</p>
          <div style="background-color: #f9f9f9; padding: 15px; border-left: 4px solid #1976d2; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>Sản phẩm:</strong> ${productName}</p>
            <p style="margin: 5px 0;"><strong>Trạng thái:</strong> ${status}</p>
            <p style="margin: 5px 0;"><strong>Hạn bảo hành:</strong> ${endDate}</p>
          </div>
          <p>Vui lòng đăng nhập vào hệ thống để xem chi tiết thông tin bảo hành mới nhất.</p>
          <p>Trân trọng,</p>
          <p>Đội ngũ ProductTrace AI</p>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }

  async sendWarrantyExpiredEmail(to: string, fullName: string, productName: string, endDate: string): Promise<void> {
    const templateId = this.configService.get<string>('WARRANTY_EXPIRED_TEMPLATE_ID') || 'd-ded28a6c91104c11bc548b08002c74f5';
    if (templateId && !templateId.includes('your_') && !templateId.includes('your-')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      
      const dynamicTemplateData = {
        fullName,
        productName,
        endDate,
        frontendUrl,
        year: new Date().getFullYear().toString(),
      };
      await this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
    } else {
      const subject = 'Thông báo hết hạn bảo hành - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #d32f2f;">Sản phẩm đã hết hạn bảo hành</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Chúng tôi xin thông báo sản phẩm <strong>${productName}</strong> của bạn đã hết hạn bảo hành vào ngày <strong>${endDate}</strong>.</p>
          <p>Nếu bạn cần hỗ trợ kỹ thuật hoặc có bất kỳ câu hỏi nào, vui lòng liên hệ với bộ phận hỗ trợ khách hàng của chúng tôi hoặc truy cập vào hệ thống để biết thêm thông tin chi tiết.</p>
          <p>Trân trọng,</p>
          <p>Đội ngũ ProductTrace AI</p>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }

  async sendOwnershipTransferredEmail(to: string, fullName: string, productName: string): Promise<void> {
    const templateId = this.configService.get<string>('OWNERSHIP_TRANSFERRED_TEMPLATE_ID') || 'd-1f78adcf1e3644e6bee66c2ea402af69';
    if (templateId && !templateId.includes('your_') && !templateId.includes('your-')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      
      const dynamicTemplateData = {
        fullName,
        productName,
        frontendUrl,
        year: new Date().getFullYear().toString(),
      };
      await this.sendEmailWithTemplate(to, templateId, dynamicTemplateData);
    } else {
      const subject = 'Thông báo chuyển quyền sở hữu sản phẩm - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #1976d2;">Thông báo chuyển quyền sở hữu</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Quyền sở hữu cho sản phẩm <strong>${productName}</strong> đã được chuyển cho bạn thành công trên hệ thống ProductTrace AI.</p>
          <p>Vui lòng đăng nhập vào hệ thống để xem chi tiết thông tin sản phẩm của bạn.</p>
          <p>Trân trọng,</p>
          <p>Đội ngũ ProductTrace AI</p>
        </div>
      `;
      await this.sendMail(to, subject, subject, html);
    }
  }
}
