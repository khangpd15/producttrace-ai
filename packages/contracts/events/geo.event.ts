// File: packages/contracts/events/geo.event.ts
//1. Payload (Dữ liệu gửi đi) của sự kiện
export interface ShopLocationSyncPayload {
  shopId: string;      // UUID của cửa hàng
  name: string;        // Tên cửa hàng
  type: 'WAREHOUSE' | 'STORE' | 'DEALER' | 'WARRANTY_CENTER'; // Loại cửa hàng
  latitude: number;    // Bắt buộc là số thực (Float/Decimal)
  longitude: number;   // Bắt buộc là số thực (Float/Decimal)
}

// 2. Routing key (Tên sự kiện) để 2 bên hứng sự kiện
export const GEO_RABBITMQ_EVENTS = {
  SHOP_LOCATION_SYNC: 'geo.shop.location_sync',
};