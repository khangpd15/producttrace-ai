const sgMail = require('@sendgrid/mail');

// BƯỚC 1: Dán SendGrid API Key của bạn vào đây (hoặc đảm bảo đã có trong file .env)
const SENDGRID_API_KEY = process.env.SENDGRID_API_KEY || 'ĐIỀN_API_KEY_CỦA_BẠN_VÀO_ĐÂY';
const SENDER_EMAIL = 'nguyenhoang280004@gmail.com'; // ĐIỀN EMAIL ĐÃ VERIFY TRÊN SENDGRID VÀO ĐÂY

sgMail.setApiKey(SENDGRID_API_KEY);

const msg = {
  to: 'nguyenhoang280004@gmail.com', // Email người nhận theo yêu cầu của bạn
  from: SENDER_EMAIL, 
  templateId: 'd-aa9b56ba4bf64b54a72eddc7ba33ba03', // ID Form của bạn
  dynamicTemplateData: {
    fullName: 'Nguyễn Hoàng',
    productName: 'iPhone 15 Pro Max 256GB',
    status: 'Đã hoàn tất bảo hành',
    endDate: '24/10/2026',
    frontendUrl: 'http://localhost:5173/warranty',
    year: new Date().getFullYear().toString(),
  },
};

console.log('Đang gửi email test đến nguyenhoang280004@gmail.com...');

sgMail
  .send(msg)
  .then(() => {
    console.log('✅ GỬI EMAIL THÀNH CÔNG! Hãy kiểm tra hộp thư nguyenhoang280004@gmail.com');
  })
  .catch((error) => {
    console.error('❌ LỖI GỬI EMAIL:');
    if (error.response) {
      console.error(error.response.body);
    } else {
      console.error(error);
    }
  });
