// File: packages/contracts/api/geo-search.dto.ts

// Dữ liệu Frontend bắt buộc phải gửi lên khi tìm kiếm
export interface GeoSearchRequestDto {
  lat: number;        // Vĩ độ hiện tại của user
  lng: number;        // Kinh độ hiện tại của user
  radiusInKm: number; // Bán kính tìm kiếm (ví dụ: 5km, 10km)
  shopType?: 'STORE' | 'WARRANTY_CENTER'; // Optional: Cần tìm cửa hàng hay trạm bảo hành
}

// Dữ liệu sẽ trả về cho Frontend
export interface GeoSearchResponseDto {
  shopId: string;
  name: string;
  type: string;
  distanceInMeters: number; // Khoảng cách từ user đến shop
  latitude: number;
  longitude: number;
}