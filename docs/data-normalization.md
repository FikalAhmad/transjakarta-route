# Stop Normalization Rules

## 1. Identifikasi Halte Fisik vs Nama Halte
Halte TransJakarta dapat memiliki nama yang sama namun merupakan physical stop berbeda (misalnya halte arah utara vs arah selatan, atau plang bus stop di seberang jalan).

## 2. Deterministic Stop Identifier Algorithm
Untuk menjaga keutuhan data dan idempotensi import, stop ID digenerate menggunakan format deterministik:

`stop_id = SHA256(stop_name + "_" + fmt.Sprintf("%.5f", lat) + "_" + fmt.Sprintf("%.5f", lon))[:16]`

Atau canonical ID:
`STP-{SLUG(stop_name)}-{LAT_LON_HASH}`

Contoh:
- `STP-BLOK_M-A1B2C3D4`
- `STP-MONAS-E5F6G7H8`

## 3. Transfer Proximity Threshold
Halte dengan nama yang sama atau halte fisik yang berjarak <= 80 meter dan melayani trip/rute berbeda dikategorikan sebagai transfer node dengan walking cost rendah (1-2 menit walking transfer penalty).
