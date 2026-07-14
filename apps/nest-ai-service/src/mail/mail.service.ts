import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { SendgridService } from './sendgrid.service';

@Injectable()
export class MailService {
  constructor(
    private readonly sendgridService: SendgridService,
    private readonly configService: ConfigService,
  ) { }

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

  async sendVerificationOtpEmail(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('VERIFY_EMAIL_TEMPLATE_ID') ||
      this.configService.get<string>('WELCOME_TEMPLATE_ID') ||
      '';

    const dynamicTemplateData = {
      fullName,
      otpCode,
      year: new Date().getFullYear().toString(),
    };

    await this.sendgridService.sendEmailWithTemplate(
      to,
      templateId,
      dynamicTemplateData
    );
  }

  async sendOTPOwnerShipEmail(to: string, productId: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('VERIFY_EMAIL_TEMPLATE_ID') ||
      this.configService.get<string>('WELCOME_TEMPLATE_ID') ||
      '';

    const dynamicTemplateData = {
      productId,
      otpCode,
      year: new Date().getFullYear().toString(),
    };

    await this.sendgridService.sendEmailWithTemplate(
      to,
      templateId,
      dynamicTemplateData
    );
  }


  async sendOTPRegister(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('VERIFY_EMAIL_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_')) {
      await this.sendVerificationOtpEmail(to, fullName, otpCode);
    } else {
      // Fallback plain email if template is not available
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
      await this.sendgridService.sendEmail(to, subject, html);
    }
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
      await this.sendgridService.sendEmail(to, subject, html);
    }
  }


  async sendOTPPasswordReset(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('SEND_EMAIL_FORGOTPASSWORD');
    if (templateId && !templateId.includes('your_')) {
      await this.sendVerificationOtpEmail(to, fullName, otpCode);
    } else {
      // Fallback plain email if template is not available
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
      await this.sendgridService.sendEmail(to, subject, html);
    }
  }

  async sendWelcomeEmail(to: string, fullName: string): Promise<void> {
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

    await this.sendgridService.sendEmailWithTemplate(
      to,
      templateId,
      dynamicTemplateData
    );
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
      await this.sendgridService.sendEmail(to, subject, html);
    }
  }

  async sendPasswordReset(to: string, fullName: string, otpCode: string): Promise<void> {
    const templateId = this.configService.get<string>('RESET_PASSWORD_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      const resetUrl = `${frontendUrl}/reset-password?email=${encodeURIComponent(to)}&code=${otpCode}`;
      await this.sendPasswordResetEmail(to, resetUrl, fullName);
    } else {
      const subject = 'Reset Your Password - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #333;">Yêu cầu đặt lại mật khẩu</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Chúng tôi nhận được yêu cầu đặt lại mật khẩu cho tài khoản của bạn.</p>
          <p>Vui lòng sử dụng mã OTP bên dưới để hoàn tất việc đặt lại mật khẩu:</p>
          <div style="background-color: #f5f5f5; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; color: #d32f2f; border-radius: 4px; margin: 20px 0;">
            ${otpCode}
          </div>
          <p>Mã này có hiệu lực trong vòng 5 phút. Nếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này.</p>
        </div>
      `;
      await this.sendgridService.sendEmail(to, subject, html);
    }
  }

  async sendWarrantyUpdateEmail(
    to: string,
    fullName: string,
    productName: string,
    status: string,
    endDate: string,
  ): Promise<void> {
    const templateId = this.configService.get<string>('WARRANTY_UPDATE_TEMPLATE_ID');
    if (templateId && !templateId.includes('your_')) {
      const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
      
      const dynamicTemplateData = {
        fullName,
        productName,
        status,
        endDate,
        frontendUrl: `${frontendUrl}/warranty`,
        year: new Date().getFullYear().toString(),
      };

      await this.sendgridService.sendEmailWithTemplate(
        to,
        templateId,
        dynamicTemplateData
      );
    } else {
      const subject = 'Cập nhật bảo hành sản phẩm - ProductTrace AI';
      const html = `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
          <h2 style="color: #1976d2;">Thông báo cập nhật bảo hành</h2>
          <p>Xin chào <strong>${fullName}</strong>,</p>
          <p>Trạng thái bảo hành cho sản phẩm <strong>${productName}</strong> của bạn vừa được cập nhật:</p>
          <p><strong>Trạng thái:</strong> ${status}</p>
          <p><strong>Hạn bảo hành:</strong> ${endDate}</p>
          <p>Vui lòng đăng nhập vào hệ thống để xem chi tiết thông tin bảo hành mới nhất.</p>
          <p>Trân trọng,</p>
          <p>Đội ngũ ProductTrace AI</p>
        </div>
      `;
      await this.sendgridService.sendEmail(to, subject, html);
    }
  }
}
