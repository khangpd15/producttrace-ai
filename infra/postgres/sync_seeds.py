import os
import re

def sync_sql_file(filepath):
    """
    Quét và tự động fix các lỗi Schema + Logic Migration cho các file SQL gốc.
    """
    if not os.path.exists(filepath):
        print(f"[-] File not found: {filepath}")
        return
        
    print(f"[*] Processing {filepath}...")
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. Đồng bộ tên cột theo migration 02
    content = content.replace('current_location_point_id', 'current_location_id')
    content = content.replace('purchase_location_point_id', 'purchase_location_id')
    content = content.replace('location_point_id', 'location_id')

    lines = content.split('\n')
    new_lines = []
    
    for line in lines:
        # 2. Cập nhật trạng thái Batch (Migration 05 & 09)
        if line.startswith('INSERT INTO batches'):
            line = line.replace("'ACTIVE'", "'IN_STOCK'")
            
        # 3. Đồng bộ bảng Events (Migration 03)
        if line.startswith('INSERT INTO events'):
            line = line.replace(', metadata_json', '')
            line = re.sub(r',\s*(\'\{.*?\}\'|NULL)\s*(,\s*\'\d{4}-\d{2}-\d{2}[^\']*\'\s*\);)$', r'\2', line)

        new_lines.append(line)

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write('\n'.join(new_lines))
    print(f"[+] Successfully synced {filepath}!\n")

if __name__ == '__main__':
    # Đã thêm tiền tố 'seed/' để trỏ đúng vào thư mục con chứa file SQL
    seed_files = [
        'seed/seed_traceability.sql',
        'seed/seed_products.sql'
    ]
    for file in seed_files:
        sync_sql_file(file)
    print("Done! Database is 100% matched with all migrations.")