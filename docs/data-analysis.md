# Data Analysis Report: TransJakarta Open Data

## 1. Schema Aktual CSV
Source: `data/raw/combine_tije_opendata_cleaned.csv`
Delimiter: `;` (Semicolon)

| Kolom | Tipe Data | Deskripsi |
|---|---|---|
| `route_id` | String | Kode rute (e.g. `1`, `10A`, `JAK.26`, `M13`) |
| `route_long_name` | String | Nama panjang rute / koridor (e.g. `Blok M - Kota`) |
| `route_color` | String (Hex) | Warna rute (e.g. `D62126`) |
| `trip_id` | String | ID perjalanan rute (e.g. `1-R07`, `1-R08`) |
| `service_id` | String | Tipe layanan / jadwal (`SH`, `HK`, `HM`, `HR`, `HL`, `TP_HK`) |
| `direction_id` | Integer (0/1) | Arah perjalanan (0 = Arah 1, 1 = Arah Sebaliknya) |
| `stop_sequence` | Integer | Urutan pemberhentian pada trip (0-indexed / 1-indexed) |
| `stop_name` | String | Nama halte / bus stop |
| `stop_lat` | Float64 | Latitude koordinat halte |
| `stop_lon` | Float64 | Longitude koordinat halte |
| `location_type` | Integer | GTFS location type (0 = stop/platform) |
| `shape_pt_sequence`| Integer | Sequence point koordinat polyline |
| `shape_id` | String | ID geometri rute (e.g. `usj3`, `mlcx`) |

## 2. Ringkasan Rekor Dataset
- **Total Baris**: 24,663
- **Baris Duplikat Eksak**: 10,777 (karena dataset digabung/flatten antara stop times & shape sequence)
- **Baris Unik Trip Stop Sequence**: 13,886
- **Unique Routes**: 245
- **Unique Trips**: 499
- **Unique Shapes**: 498
- **Unique Stop Names**: 6,832
- **Missing / Null Values**: 0 pada semua kolom utama

## 3. Geographic Bounds
- **Latitude Range**: `[-6.402636, -6.082729]` (Wilayah Jabodetabek / DKI Jakarta)
- **Longitude Range**: `[106.671391, 107.018981]`

## 4. Analisis Struktur & Relasi
1. **Route → Trip**: Setiap Route memiliki 1 atau lebih Trip (rata-rata 2 trip per rute untuk direction 0 dan direction 1).
2. **Trip → Stops**: Trip memiliki urutan halte (`stop_sequence`).
3. **Stop Coordinates**: Koordinat latitude/longitude konsisten dengan sistem WGS84 (SRID 4326).
4. **Service Classification**:
   - `BRT`: Rute Koridor Utama (Rute 1 sampai 14 numerik murni).
   - `NON_BRT` / Feeder / Mikrotrans: Rute dengan awalan `JAK.` (Mikrotrans), `M` (Minitrans/Angkutan Malam), `[0-9]+[A-Z]` (Feeder), dll.

## 5. Keputusan Normalisasi & Pipeline
1. **Idempotent Ingestion**: Data duplikat difilter saat parsing berdasarkan `(trip_id, stop_sequence, stop_name)`.
2. **Stop Identity**: Halte fisik diidentifikasi deterministik menggunakan hash dari `(stop_name, rounded(lat, 5), rounded(lon, 5))` agar tidak rancu dengan nama halte yang sama di dua sisi jalan yang berbeda.
3. **Graph Construction**:
   - Setiap trip diurutkan berdasarkan `stop_sequence` menaik.
   - Transit edge dibuat dari `Stop[i] → Stop[i+1]` dengan atribut trip, route, direction.
   - Transfer edge dibuat antar rute berbeda yang bersinggungan di stop fisik yang sama atau radius halte terhubung (< 80m).
