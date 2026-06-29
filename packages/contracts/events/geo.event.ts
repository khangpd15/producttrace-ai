// File: packages/contracts/events/geo.event.ts
//1. Payload (Dữ liệu gửi đi) của sự kiện
export interface ShopLocationSyncPayload {
  shopId: string;      // UUID của cửa hàng
  name: string;        // Tên cửa hàng
  type: 'WAREHOUSE' | 'STORE' | 'DEALER' | 'WARRANTY_CENTER'; // Loại cửa hàng
  latitude: number;    // Vĩ độ (Float/Decimal)
  longitude: number;   // Kinh độ (Float/Decimal)
}

// 2. Routing key để 2 bên hứng sự kiện
export const GEO_RABBITMQ_EVENTS = {
  SHOP_LOCATION_SYNC: 'geo.shop.location_sync',
};