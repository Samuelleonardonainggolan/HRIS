"""
=============================================================================
Hotel Labersa Toba — Stateless Design
=============================================================================
Prinsip:
  - TIDAK ada database, TIDAK ada auth
  - Terima request dari Golang → proses → kembalikan hasil
  - Embedding disimpan di Golang, dikirim ke sini saat verifikasi
  - Dilindungi API Key (hanya Golang yang boleh akses)

Endpoint:
  POST /face/extract         ← Ekstrak embedding dari foto (saat registrasi)
  POST /face/verify          ← Cocokkan foto vs embedding acuan (saat absen)
  POST /geo/validate         ← Validasi koordinat GPS
  POST /attendance/process   ← Pipeline lengkap: geo + face sekaligus
  GET  /health               ← Status service
=============================================================================
"""
"import get new data from golang to save the database and push data for facenet to process the data and retur to get data "

import io
import os
import json
import math
import time
import logging
from typing import Optional

import numpy as np
import torch
import torch.nn as nn
from fastapi import FastAPI, File, UploadFile, Form, Header, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from PIL import Image
from torchvision import transforms
from contextlib import asynccontextmanager

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger("face-service")

# =============================================================================
# KONFIGURASI
# =============================================================================
MODEL_PATH          = os.getenv("MODEL_PATH", r"D:\Dataset\output_cpu\facenet_labersa_cpu.pt")
IMAGE_SIZE          = int(os.getenv("IMAGE_SIZE", "160"))
SIMILARITY_THRESHOLD = float(os.getenv("SIMILARITY_THRESHOLD", "0.75"))
OFFICE_LAT          = float(os.getenv("OFFICE_LAT", "2.6559"))
OFFICE_LNG          = float(os.getenv("OFFICE_LNG", "98.9003"))
GEOFENCE_RADIUS_M   = float(os.getenv("GEOFENCE_RADIUS_M", "100"))
API_KEY             = os.getenv("FACE_API_KEY", "labersa-internal-api-key-2026")

# =============================================================================
# MODEL
# =============================================================================
class FaceNetExtractor(nn.Module):
    def __init__(self):
        super().__init__()
        from facenet_pytorch import InceptionResnetV1
        self.backbone = InceptionResnetV1(pretrained=None, classify=False)
        for p in self.backbone.parameters():
            p.requires_grad = False

    def forward(self, x):
        return nn.functional.normalize(self.backbone(x), p=2, dim=1)


class LightClassifier(nn.Module):
    def __init__(self, num_classes: int, dropout: float = 0.5):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(512, 256), nn.BatchNorm1d(256), nn.GELU(), nn.Dropout(dropout),
            nn.Linear(256, 128), nn.BatchNorm1d(128), nn.GELU(), nn.Dropout(dropout * 0.6),
            nn.Linear(128, num_classes),
        )

    def forward(self, x):
        return self.net(x)


# =============================================================================
# FACE SERVICE (stateless)
# =============================================================================
class FaceService:
    def __init__(self):
        self.extractor   = None
        self.classifier  = None
        self.class_names = []
        self.loaded      = False
        self.transform   = transforms.Compose([
            transforms.Resize((IMAGE_SIZE, IMAGE_SIZE)),
            transforms.ToTensor(),
            transforms.Normalize([0.5]*3, [0.5]*3),
        ])

    def load(self):
        try:
            self.extractor = FaceNetExtractor().eval()

            if os.path.exists(MODEL_PATH):
                ckpt = torch.load(MODEL_PATH, map_location="cpu")
                self.class_names = ckpt.get("class_names", [])
                if "classifier_state_dict" in ckpt and self.class_names:
                    self.classifier = LightClassifier(len(self.class_names)).eval()
                    self.classifier.load_state_dict(ckpt["classifier_state_dict"])
                logger.info(f"Model loaded: {len(self.class_names)} classes")
            else:
                # Fallback: pakai pretrained weights langsung
                from facenet_pytorch import InceptionResnetV1
                self.extractor.backbone = InceptionResnetV1(
                    pretrained="vggface2", classify=False
                ).eval()
                logger.warning(f"Model tidak ditemukan di {MODEL_PATH}, pakai pretrained VGGFace2")

            self.loaded = True
        except Exception as e:
            logger.error(f"Gagal load model: {e}")
            raise

    @torch.no_grad()
    def extract_embedding(self, image_bytes: bytes) -> list[float]:
        """
        Ekstrak 512-dim embedding dari foto wajah.
        Return: list[float] — dikirim ke Golang untuk disimpan di DB-nya.
        """
        img    = Image.open(io.BytesIO(image_bytes)).convert("RGB")
        tensor = self.transform(img).unsqueeze(0)
        emb    = self.extractor(tensor)
        return emb[0].cpu().numpy().tolist()   # list 512 float

    def cosine_similarity(self, emb1: list, emb2: list) -> float:
        a = np.array(emb1, dtype=np.float32)
        b = np.array(emb2, dtype=np.float32)
        a /= (np.linalg.norm(a) + 1e-8)
        b /= (np.linalg.norm(b) + 1e-8)
        return float(np.dot(a, b))

    @torch.no_grad()
    def verify(
        self,
        image_bytes     : bytes,
        stored_embedding: list[float],
        threshold       : float = None,
    ) -> dict:
        """
        Bandingkan foto wajah vs embedding acuan yang dikirim dari Golang.

        Args:
            image_bytes      : Foto wajah saat absen (bytes)
            stored_embedding : Embedding acuan (list 512 float, dari Golang DB)
            threshold        : Override threshold (opsional)

        Returns:
            dict berisi matched, similarity, confidence, message
        """
        thr = threshold or SIMILARITY_THRESHOLD

        live_emb   = self.extract_embedding(image_bytes)
        similarity = self.cosine_similarity(live_emb, stored_embedding)
        matched    = similarity >= thr

        # Confidence dari classifier jika tersedia
        confidence = similarity
        if self.classifier and self.class_names:
            img    = Image.open(io.BytesIO(image_bytes)).convert("RGB")
            tensor = self.transform(img).unsqueeze(0)
            emb_t  = self.extractor(tensor)
            logits = self.classifier(emb_t)
            probs  = torch.softmax(logits, dim=1)[0]
            confidence = float(probs.max().item())

        return {
            "matched"   : matched,
            "similarity": round(similarity, 4),
            "confidence": round(confidence, 4),
            "threshold" : thr,
            "message"   : (
                f"Wajah cocok (similarity={similarity:.1%})"
                if matched else
                f"Wajah tidak cocok (similarity={similarity:.1%}, min={thr:.0%})"
            ),
        }


face_svc = FaceService()


# =============================================================================
# GEOFENCING
# =============================================================================
def haversine(lat1, lng1, lat2, lng2) -> float:
    R    = 6_371_000
    phi1 = math.radians(lat1); phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlam = math.radians(lng2 - lng1)
    a    = math.sin(dphi/2)**2 + math.cos(phi1)*math.cos(phi2)*math.sin(dlam/2)**2
    return 2 * R * math.atan2(math.sqrt(a), math.sqrt(1-a))


def validate_location(latitude: float, longitude: float,
                      radius_m: float = None) -> dict:
    radius   = radius_m or GEOFENCE_RADIUS_M
    distance = haversine(latitude, longitude, OFFICE_LAT, OFFICE_LNG)
    valid    = distance <= radius
    return {
        "is_valid"   : valid,
        "distance_m" : round(distance, 1),
        "radius_m"   : radius,
        "office_lat" : OFFICE_LAT,
        "office_lng" : OFFICE_LNG,
        "message"    : (
            f"Lokasi valid, jarak {distance:.0f}m dari kantor"
            if valid else
            f"Diluar radius, jarak {distance:.0f}m (maks {radius:.0f}m)"
        ),
    }


# =============================================================================
# SCHEMAS
# =============================================================================
class VerifyRequest(BaseModel):
    """
    Request body untuk verifikasi wajah.
    stored_embedding dikirim dari Golang (diambil dari DB Golang).
    """
    stored_embedding: list[float]
    employee_id     : str
    threshold       : Optional[float] = None


class GeoRequest(BaseModel):
    latitude : float
    longitude: float
    radius_m : Optional[float] = None


class AttendanceProcessRequest(BaseModel):
    """
    Pipeline lengkap — Golang kirim semua data sekaligus,
    FastAPI kembalikan keputusan akhir.
    """
    employee_id      : str
    stored_embedding : list[float]
    latitude         : float
    longitude        : float
    record_type      : str             # "checkin" / "checkout"
    threshold        : Optional[float] = None
    radius_m         : Optional[float] = None


# =============================================================================
# SECURITY — API KEY
# =============================================================================
def verify_api_key(x_api_key: str = Header(..., alias="X-API-Key")):
    """
    Semua endpoint dilindungi API Key.
    Golang menyertakan header: X-API-Key: <key>
    """
    if x_api_key != API_KEY:
        raise HTTPException(401, "API Key tidak valid")
    return x_api_key


# =============================================================================
# APP
# =============================================================================
@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Memuat model FaceNet...")
    face_svc.load()
    logger.info("✅ Face Recognition Service siap")
    yield
    logger.info("Service berhenti.")


app = FastAPI(
    title      = "Face Recognition  — Hotel Labersa Toba",
    description= """
**Internal  untuk face recognition & geofencing.
Dipanggil oleh Golang Backend, bukan langsung oleh client.

### Alur:
1. **Registrasi wajah**: Golang kirim foto → FastAPI ekstrak embedding → 
   kembalikan `embedding[]` → Golang simpan di DB-nya
2. **Absensi**: Golang ambil embedding dari DB → kirim ke FastAPI bersama foto baru → 
   FastAPI bandingkan → kembalikan `matched: true/false`
    """,
    version    = "1.0.0",
    lifespan   = lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["POST", "GET"],
    allow_headers=["*"],
)


# =============================================================================
# ENDPOINTS
# =============================================================================

# ── Health Check ──────────────────────────────────────────────────────────────
@app.get("/health", summary="Status service")
def health():
    return {
        "status"       : "ok",
        "model_loaded" : face_svc.loaded,
        "num_classes"  : len(face_svc.class_names),
        "office_coords": {"lat": OFFICE_LAT, "lng": OFFICE_LNG},
        "geofence_m"   : GEOFENCE_RADIUS_M,
        "threshold"    : SIMILARITY_THRESHOLD,
    }


# ── 1. Ekstrak Embedding (saat registrasi wajah) ──────────────────────────────
@app.post(
    "/face/extract",
    summary="📸 Ekstrak embedding dari foto wajah",
    description="""
**Dipakai saat registrasi wajah pegawai.**

Golang mengirim foto wajah → FastAPI mengekstrak 512-dim embedding →
embedding dikembalikan ke Golang untuk disimpan di database Golang.

**Request**: `multipart/form-data`
- `photo`: file gambar (JPG/PNG)
- `employee_id`: ID pegawai (untuk logging)

**Response**: `embedding` berupa array 512 float yang harus disimpan Golang.
    """
)
async def extract_embedding(
    photo      : UploadFile = File(..., description="Foto wajah (JPG/PNG/WEBP)"),
    employee_id: str        = Form(..., description="ID pegawai untuk logging"),
    _          = Depends(verify_api_key),
):
    _validate_image_file(photo)
    image_bytes = await photo.read()

    t0        = time.time()
    embedding = face_svc.extract_embedding(image_bytes)
    elapsed   = round((time.time() - t0) * 1000, 1)

    logger.info(f"[EXTRACT] employee={employee_id} elapsed={elapsed}ms")

    return {
        "success"    : True,
        "employee_id": employee_id,
        "embedding"  : embedding,          # ← Golang simpan ini di DB
        "dimension"  : len(embedding),     # selalu 512
        "elapsed_ms" : elapsed,
        "message"    : "Embedding berhasil diekstrak. Simpan nilai 'embedding' di database Anda.",
    }


# ── 2. Verifikasi Wajah (saat absensi) ───────────────────────────────────────
@app.post(
    "/face/verify",
    summary="🔍 Cocokkan foto vs embedding acuan",
    description="""
**Dipakai saat pegawai absen (check-in / check-out).**

Alur di sisi Golang:
1. Golang ambil `stored_embedding` pegawai dari database Golang
2. Golang kirim foto selfie + stored_embedding ke endpoint ini
3. FastAPI bandingkan → kembalikan `matched: true/false`
4. Golang catat hasil absensi ke database-nya

**Request**: `multipart/form-data`
- `photo`: foto selfie saat absen
- `data`: JSON string berisi `employee_id`, `stored_embedding`, `threshold` (opsional)
    """
)
async def verify_face(
    photo: UploadFile = File(..., description="Foto selfie saat absen"),
    data : str        = Form(..., description='JSON: {"employee_id":"...","stored_embedding":[...],"threshold":0.75}'),
    _    = Depends(verify_api_key),
):
    # Parse JSON body dari form field
    try:
        req = VerifyRequest(**json.loads(data))
    except Exception as e:
        raise HTTPException(400, f"Format 'data' tidak valid: {e}")

    if len(req.stored_embedding) != 512:
        raise HTTPException(400, f"stored_embedding harus 512 dimensi, dapat {len(req.stored_embedding)}")

    _validate_image_file(photo)
    image_bytes = await photo.read()

    t0     = time.time()
    result = face_svc.verify(image_bytes, req.stored_embedding, req.threshold)
    elapsed = round((time.time() - t0) * 1000, 1)

    logger.info(
        f"[VERIFY] employee={req.employee_id} "
        f"matched={result['matched']} sim={result['similarity']:.3f} "
        f"elapsed={elapsed}ms"
    )

    return {
        **result,
        "employee_id": req.employee_id,
        "elapsed_ms" : elapsed,
    }


# ── 3. Validasi GPS ───────────────────────────────────────────────────────────
@app.post(
    "/geo/validate",
    summary="📍 Validasi koordinat GPS",
    description="""
Cek apakah koordinat pegawai berada dalam radius kantor.

Bisa dipanggil terpisah dari verifikasi wajah, atau gunakan
`/attendance/process` untuk keduanya sekaligus.
    """
)
def validate_geo(
    req: GeoRequest,
    _  = Depends(verify_api_key),
):
    result = validate_location(req.latitude, req.longitude, req.radius_m)
    logger.info(f"[GEO] lat={req.latitude} lng={req.longitude} valid={result['is_valid']} dist={result['distance_m']}m")
    return result


# ── 4. Pipeline Lengkap — ENDPOINT UTAMA ─────────────────────────────────────
@app.post(
    "/attendance/process",
    summary="✅ Pipeline lengkap: GPS + Face Verification",
    description="""
**Endpoint utama untuk absensi.** Golang cukup panggil satu endpoint ini.

Proses internal:
1. Validasi GPS → jika gagal, langsung return tanpa proses wajah
2. Verifikasi wajah → cocokkan foto vs embedding acuan
3. Kembalikan keputusan akhir beserta detail

**Request**: `multipart/form-data`
- `photo` : foto selfie
- `data`  : JSON string semua parameter

**Response** berisi `decision`:
- `approved`     → boleh absen (GPS ✓ + wajah ✓)
- `rejected_gps` → di luar radius kantor
- `rejected_face`→ wajah tidak cocok
    """
)
async def process_attendance(
    photo: UploadFile = File(..., description="Foto selfie"),
    data : str        = Form(..., description="""JSON: {
  "employee_id": "uuid-dari-golang",
  "stored_embedding": [...512 float...],
  "latitude": 2.6559,
  "longitude": 98.9003,
  "record_type": "checkin",
  "threshold": 0.75,
  "radius_m": 100
}"""),
    _    = Depends(verify_api_key),
):
    try:
        req = AttendanceProcessRequest(**json.loads(data))
    except Exception as e:
        raise HTTPException(400, f"Format 'data' tidak valid: {e}")

    if len(req.stored_embedding) != 512:
        raise HTTPException(400, f"stored_embedding harus 512 dimensi, dapat {len(req.stored_embedding)}")

    _validate_image_file(photo)
    image_bytes = await photo.read()

    t0 = time.time()

    # ── Step 1: Validasi GPS ──────────────────────────────────────────────
    geo = validate_location(req.latitude, req.longitude, req.radius_m)

    if not geo["is_valid"]:
        elapsed = round((time.time() - t0) * 1000, 1)
        logger.info(
            f"[ATTENDANCE] employee={req.employee_id} type={req.record_type} "
            f"decision=rejected_gps dist={geo['distance_m']}m elapsed={elapsed}ms"
        )
        return {
            "decision"      : "rejected_gps",
            "approved"      : False,
            "employee_id"   : req.employee_id,
            "record_type"   : req.record_type,
            "geo"           : geo,
            "face"          : None,
            "elapsed_ms"    : elapsed,
            "message"       : geo["message"],
        }

    # ── Step 2: Verifikasi Wajah ──────────────────────────────────────────
    face = face_svc.verify(image_bytes, req.stored_embedding, req.threshold)

    elapsed  = round((time.time() - t0) * 1000, 1)
    approved = face["matched"]
    decision = "approved" if approved else "rejected_face"

    logger.info(
        f"[ATTENDANCE] employee={req.employee_id} type={req.record_type} "
        f"decision={decision} sim={face['similarity']:.3f} "
        f"dist={geo['distance_m']}m elapsed={elapsed}ms"
    )

    return {
        "decision"   : decision,
        "approved"   : approved,
        "employee_id": req.employee_id,
        "record_type": req.record_type,
        "geo"        : geo,
        "face"       : face,
        "elapsed_ms" : elapsed,
        "message"    : (
            f"Absensi {req.record_type} disetujui"
            if approved else face["message"]
        ),
    }


# =============================================================================
# HELPER
# =============================================================================
ALLOWED_EXT = {"jpg", "jpeg", "png", "webp"}

def _validate_image_file(file: UploadFile):
    ext = (file.filename or "").rsplit(".", 1)[-1].lower()
    if ext not in ALLOWED_EXT:
        raise HTTPException(400, f"Format tidak didukung. Gunakan: {ALLOWED_EXT}")


# =============================================================================
# RUN
# =============================================================================
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host    = "0.0.0.0",
        port    = int(os.getenv("PORT", "8001")),
        reload  = False,
        workers = 1,   # 1 worker — model FaceNet tidak thread-safe
    )
