import json

with open(r'D:\Project\omni\notebooks\tiktok_ads\data\product_names.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

tenant = 'yumna_bertigamart'
schema = f'tenant_{tenant}'

lines = []
lines.append(f'SET search_path TO {schema}, public;')
lines.append('')
lines.append('CREATE TABLE IF NOT EXISTS tiktok_ads_product_names (')
lines.append('    id SERIAL PRIMARY KEY,')
lines.append('    tenant_id VARCHAR(255) NOT NULL,')
lines.append('    product_id VARCHAR(255) NOT NULL,')
lines.append('    name TEXT,')
lines.append('    category VARCHAR(500),')
lines.append('    created_at TIMESTAMPTZ DEFAULT NOW(),')
lines.append('    updated_at TIMESTAMPTZ DEFAULT NOW()')
lines.append(');')
lines.append('')
lines.append(f"DELETE FROM tiktok_ads_product_names WHERE tenant_id = '{tenant}';")
lines.append('')

for pid, info in data.items():
    name = info['name'].replace("'", "''")
    cat = info.get('category', '').replace("'", "''")
    lines.append(
        f"INSERT INTO tiktok_ads_product_names (tenant_id, product_id, name, category, created_at, updated_at) "
        f"VALUES ('{tenant}', '{pid}', '{name}', '{cat}', NOW(), NOW());"
    )

lines.append('')
lines.append('SELECT COUNT(*) as total_imported FROM tiktok_ads_product_names;')

sql = '\n'.join(lines)
with open(r'D:\Project\omni\scripts\import_product_names.sql', 'w', encoding='utf-8') as f:
    f.write(sql)

print(f'Generated SQL for {len(data)} products -> scripts/import_product_names.sql')
