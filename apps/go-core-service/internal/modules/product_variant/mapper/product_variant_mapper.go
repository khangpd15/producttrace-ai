package mapper

import (
    "encoding/json"

    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/dto/response"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
)

func ToVariantResponse(variant *entities.ProductVariant) *response.VariantResponse {
    var images []string
    if variant.ImagesJSON != nil {
        json.Unmarshal(variant.ImagesJSON, &images)
    }

    return &response.VariantResponse{
        ID:        variant.ID,
        ProductID: variant.ProductID,
        SKU:       variant.SKU,
        Name:      variant.Name,
        Barcode:   variant.Barcode,
        Price:     variant.Price,
        Currency:  variant.Currency,
        Images:    images,
        Status:    variant.Status,
        CreatedAt: variant.CreatedAt,
        UpdatedAt: variant.UpdatedAt,
    }
}

func ToListVariantResponse(variants []entities.ProductVariant, total int64) *response.ListVariantResponse {
    data := make([]response.VariantResponse, len(variants))
    for i, v := range variants {
        data[i] = *ToVariantResponse(&v)
    }
    return &response.ListVariantResponse{
        Data:  data,
        Total: total,
    }
}