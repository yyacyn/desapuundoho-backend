# Puundoho API Documentation

## Overview

A collection of API endpoints for the **Puundoho** application, covering articles, gallery, listings, and population statistics.

**Base URL:** `http://localhost:8081`

---

## Variables

| Variable | Value           |
|----------|-----------------|
| `local`  | `localhost:8081` |

---

## Endpoints

### 1. GET Articles

Retrieve a list of published articles.

- **Method:** `GET`
- **URL:** `http://{{local}}/api/articles`
- **Authentication:** None

#### Response — `200 OK`

```json
{
  "articles": [
    {
      "id": 2,
      "title": "pochita",
      "slug": "pochita",
      "content": "<p>aaaaaaa</p>",
      "excerpt": "dream on",
      "cover_image": "https://ik.imagekit.io/yyacyn/articles/ec705c039692f211296e621444ff7764_alW23V9xh.jpg",
      "author": "admin",
      "category": "ssss",
      "status": "published",
      "created_at": "2026-03-17T08:31:33.740874Z",
      "updated_at": "2026-03-17T08:34:10.121014Z"
    },
    {
      "id": 1,
      "title": "mai",
      "slug": "mai",
      "content": "<p>adsasda</p>",
      "excerpt": "uayay",
      "cover_image": "https://ik.imagekit.io/yyacyn/articles/c178fe4c2b6a4f495a3e0d461af69dac_PphN06myA.jpg",
      "author": "admin",
      "category": "hmmm",
      "status": "published",
      "created_at": "2026-03-17T08:23:36.323874Z",
      "updated_at": "2026-03-17T08:23:36.323874Z"
    }
  ]
}
```

#### Response Fields

| Field        | Type     | Description                              |
|--------------|----------|------------------------------------------|
| `articles`   | array    | List of article objects                  |
| `id`         | int      | Unique identifier of the article         |
| `title`      | string   | Title of the article                     |
| `slug`       | string   | URL-friendly identifier                  |
| `content`    | string   | HTML content of the article              |
| `excerpt`    | string   | Short summary of the article             |
| `cover_image`| string   | URL of the article's cover image         |
| `author`     | string   | Author of the article                    |
| `category`   | string   | Category the article belongs to          |
| `status`     | string   | Publication status (e.g. `published`)    |
| `created_at` | string   | ISO 8601 creation timestamp              |
| `updated_at` | string   | ISO 8601 last updated timestamp          |

---

### 2. GET Galeri

Retrieve a list of gallery items with images.

- **Method:** `GET`
- **URL:** `http://{{local}}/api/galeri`
- **Authentication:** None

#### Response — `200 OK`

```json
{
  "galeri": [
    {
      "id": 1,
      "images": [
        "https://ik.imagekit.io/yyacyn/galeri/44e2c88961067bde33a27c6c54699a8c_57Z8gHXjjN.jpg",
        "https://ik.imagekit.io/yyacyn/galeri/WhatsApp_Image_2026-03-12_at_14.06.00_BPpjcwVi2.jpeg"
      ],
      "caption": "seiba",
      "created_at": "2026-03-17T08:57:21.910822Z"
    }
  ]
}
```

#### Response Fields

| Field        | Type   | Description                              |
|--------------|--------|------------------------------------------|
| `galeri`     | array  | List of gallery objects                  |
| `id`         | int    | Unique identifier of the gallery item    |
| `images`     | array  | List of image URLs in the gallery item   |
| `caption`    | string | Caption for the gallery item             |
| `created_at` | string | ISO 8601 creation timestamp              |

---

### 3. GET Listings

Retrieve a list of location listings.

- **Method:** `GET`
- **URL:** `http://{{local}}/api/listings`
- **Authentication:** None

#### Response — `200 OK`

```json
{
  "listings": [
    {
      "id": 1,
      "nama": "ougi",
      "koordinat": "-3.11131, 121.08936",
      "image_url": "https://ik.imagekit.io/yyacyn/listings/G6cdDgHbwAAEBT6_ZUoJQ6ops.jpg",
      "created_at": "2026-03-16T15:55:29.418643Z"
    }
  ]
}
```

#### Response Fields

| Field        | Type   | Description                              |
|--------------|--------|------------------------------------------|
| `listings`   | array  | List of listing objects                  |
| `id`         | int    | Unique identifier of the listing         |
| `nama`       | string | Name of the listing                      |
| `koordinat`  | string | Latitude and longitude coordinates       |
| `image_url`  | string | URL of the listing's image               |
| `created_at` | string | ISO 8601 creation timestamp              |

---

### 4. GET Population Stats

Retrieve population statistics for a given dataset.

- **Method:** `GET`
- **URL:** `http://{{local}}/api/penduduk/datasets/1/stats`
- **Authentication:** None

> **Note:** Replace `1` in the URL with the desired `dataset_id`.

#### Response — `200 OK`

```json
{
  "age_range": {
    "0-5": 70,
    "6-12": 131,
    "13-17": 109,
    "18-59": 729,
    "60+": 100
  },
  "dusun": {
    "Dusun 1": 464,
    "Dusun 2": 211,
    "Dusun 3": 382,
    "Dusun 4": 78,
    "Dusun 5": 1,
    "Tidak Diketahui": 3
  },
  "education": {
    "Belum/Tidak Sekolah": 196,
    "SD Sederajat": 390,
    "SMP Sederajat": 180,
    "SMA Sederajat": 277,
    "D3": 22,
    "D4": 2,
    "S1": 71,
    "Tidak Diketahui": 1
  },
  "gender": {
    "Laki-laki": 571,
    "Perempuan": 568
  },
  "job": {
    "ASN/TNI/POLRI": 27,
    "Belum/Tidak Bekerja": 291,
    "Honorer": 19,
    "IRT": 251,
    "Karyawan": 13,
    "Nelayan": 1,
    "Pelajar/Mahasiswa": 255,
    "Pelaut": 1,
    "Pensiunan": 2,
    "Perangkat Desa": 2,
    "Perawat": 3,
    "Petani": 211,
    "Sopir": 2,
    "Tidak Diketahui": 4,
    "Tukang": 2,
    "Wiraswasta": 55
  },
  "marriage": {
    "Belum Kawin": 569,
    "Kawin": 509,
    "Cerai Hidup": 21,
    "Cerai Mati": 38,
    "Tidak Diketahui": 2
  },
  "religion": {
    "Islam": 1123,
    "Kristen": 15,
    "Tidak Diketahui": 1
  }
}
```

#### Response Fields

| Field        | Type   | Description                                      |
|--------------|--------|--------------------------------------------------|
| `age_range`  | object | Population count grouped by age range            |
| `dusun`      | object | Population count grouped by village (dusun)      |
| `education`  | object | Population count grouped by education level      |
| `gender`     | object | Population count grouped by gender               |
| `job`        | object | Population count grouped by occupation           |
| `marriage`   | object | Population count grouped by marital status       |
| `religion`   | object | Population count grouped by religion             |