# Service Classification Documentation

## 1. Kategori Layanan TransJakarta

Berdasarkan analisis aktual rute pada dataset `combine_tije_opendata_cleaned.csv`:

### A. BRT (Bus Rapid Transit - Koridor Utama)
- **Karakteristik ID**: Angka murni dari 1 hingga 14 (misal: `1`, `2`, `3`, ..., `13`, `14`).
- **Jalur**: Khusus busway / halte bertingkat & tertutup dengan platform tinggi.
- **Karakteristik Operasional**: Jalur steril, headway cepat, kapasitas tinggi.

### B. Feeder / Non-BRT (Bus Kota / Integrasi Pemukiman / Rusun)
- **Karakteristik ID**: Angka diikuti huruf (misal: `1A`, `6H`, `7A`, `10A`, `11D`, `5M`, `4B`, dll).
- **Jalur**: Campuran jalur busway dan jalan reguler, berhenti di bus stop / plang halte.
- **Karakteristik Operasional**: Menghubungkan stasiun KRL / MRT / perumahan ke koridor utama BRT.

### C. Mikrotrans (JakLingko Angkot)
- **Karakteristik ID**: Menggunakan prefix `JAK.` (misal: `JAK.01`, `JAK.26`, `JAK.108`, dll).
- **Jalur**: Jalan lingkungan dan arteri sekunder.
- **Karakteristik Operasional**: Angkutan first-mile / last-mile gratis dengan kartu JakLingko.

### D. Angkutan Malam Hari (AMARI) & Lainnya
- **Karakteristik ID**: Prefix `M` (misal: `M13`, `M5`, `M11`).
- **Jalur**: Beroperasi pada koridor tertentu saat malam hari.

## 2. Implementasi Rule Classification di Go
```go
func ClassifyService(routeID string) string {
    if _, err := strconv.Atoi(routeID); err == nil {
        return "BRT"
    }
    if strings.HasPrefix(routeID, "JAK.") {
        return "MIKROTRANS"
    }
    if strings.HasPrefix(routeID, "M") {
        return "AMARI"
    }
    return "NON_BRT"
}
```
