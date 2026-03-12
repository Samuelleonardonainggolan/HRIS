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

import io
import os
import json
import math
import time
import logging
from typing import Optional, Tuple

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
from facenet_pytorch import MTCNN, InceptionResnetV1

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger("face-service")

# =============================================================================
# KONFIGURASI
# =============================================================================
MODEL_PATH = os.getenv("MODEL_PATH", r"D:\Dataset\output_cpu\facenet_labersa_cpu.pt")
IMAGE_SIZE = int(os.getenv("IMAGE_SIZE", "160"))
SIMILARITY_THRESHOLD = float(os.getenv("SIMILARITY_THRESHOLD", "0.75"))
OFFICE_LAT = float(os.getenv("OFFICE_LAT", "2.6559"))
OFFICE_LNG = float(os.getenv("OFFICE_LNG", "98.9003"))
GEOFENCE_RADIUS_M = float(os.getenv("GEOFENCE_RADIUS_M", "100"))
API_KEY = os.getenv("FACE_API_KEY", "labersa-internal-api-key-2026")

# =============================================================================
# MODEL
# =============================================================================
class FaceNetExtractor(nn.Module):
    def __init__(self):
        super().__init__()
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
        self.extractor = None
        self.classifier = None
        self.class_names = []
        self.loaded = False
        self.mtcnn = None
        self.transform = transforms.Compose([
            transforms.Resize((IMAGE_SIZE, IMAGE_SIZE)),
            transforms.ToTensor(),
            transforms.Normalize([0.5]*3, [0.5]*3),
        ])

    def load(self):
        try:
            # Inisialisasi MTCNN untuk deteksi wajah
            self.mtcnn = MTCNN(keep_all=True)
            logger.info("MTCNN loaded successfully")

            # Inisialisasi extractor
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
                self.extractor.backbone = InceptionResnetV1(
                    pretrained="vggface2", classify=False
                ).eval()
                logger.warning(f"Model tidak ditemukan di {MODEL_PATH}, pakai pretrained VGGFace2")

            self.loaded = True
        except Exception as e:
            logger.error(f"Gagal load model: {e}")
            raise

    def detect_faces(self, image_bytes: bytes) -> Tuple[bool, int, list]:
        """
        Deteksi wajah dalam gambar menggunakan MTCNN
        Returns: (has_face, face_count, boxes)
        """
        try:
            if self.mtcnn is None:
                self.mtcnn = MTCNN(keep_all=True)

            img = Image.open(io.BytesIO(image_bytes)).convert("RGB")
            boxes, _ = self.mtcnn.detect(img)

            if boxes is None or len(boxes) == 0:
                return False, 0, []

            # Validasi ukuran wajah (minimal 50x50 pixel)
            valid_boxes = []
            for box in boxes:
                width = box[2] - box[0]
                height = box[3] - box[1]
                if width >= 50 and height >= 50:
                    valid_boxes.append(box)

            return len(valid_boxes) > 0, len(valid_boxes), valid_boxes
        except Exception as e:
            logger.error(f"Error detecting faces: {e}")
            return False, 0, []

    def check_accessories(self, image_bytes: bytes, boxes: list) -> Tuple[bool, str]:
        """
        Periksa apakah ada aksesoris yang menutupi wajah
        Returns: (is_valid, message)
        """
        try:
            img = Image.open(io.BytesIO(image_bytes)).convert("RGB")
            img_array = np.array(img)
            
            # Ambil wajah pertama untuk diperiksa
            if len(boxes) == 0:
                return False, "Tidak ada wajah terdeteksi"
            
            box = boxes[0]
            x1, y1, x2, y2 = [int(b) for b in box]
            
            # Pastikan koordinat dalam batas gambar
            x1 = max(0, x1)
            y1 = max(0, y1)
            x2 = min(img_array.shape[1], x2)
            y2 = min(img_array.shape[0], y2)
            
            face_region = img_array[y1:y2, x1:x2]
            
            if face_region.size == 0:
                return False, "Region wajah tidak valid"
            
            # Konversi ke HSV untuk analisis warna
            hsv_face = cv2.cvtColor(face_region, cv2.COLOR_RGB2HSV)
            
            # Deteksi kacamata hitam (area gelap di sekitar mata)
            # Sederhana: cek rasio pixel gelap di setengah atas wajah
            height, width = face_region.shape[:2]
            upper_half = face_region[0:height//2, :]
            
            # Hitung kecerahan rata-rata
            brightness = np.mean(cv2.cvtColor(upper_half, cv2.COLOR_RGB2GRAY))
            
            # Jika terlalu gelap, kemungkinan kacamata hitam
            if brightness < 50:
                return False, "Terdeteksi kacamata hitam, harap lepas"
            
            # Deteksi masker (area gelap di sekitar mulut)
            # Sederhana: cek rasio pixel gelap di sepertiga bawah wajah
            lower_third = face_region[2*height//3:height, :]
            lower_brightness = np.mean(cv2.cvtColor(lower_third, cv2.COLOR_RGB2GRAY))
            
            # Jika terlalu gelap, kemungkinan masker
            if lower_brightness < 40:
                return False, "Terdeteksi masker, harap lepas"
            
            # Deteksi topi/aksesoris kepala (cek bagian atas wajah)
            top_part = face_region[0:height//4, :]
            if np.mean(cv2.cvtColor(top_part, cv2.COLOR_RGB2GRAY)) < 30:
                return False, "Terdeteksi aksesoris di kepala, harap lepas"
            
            return True, "Wajah valid"
        except Exception as e:
            logger.error(f"Error checking accessories: {e}")
            return True, "Tidak dapat memeriksa aksesoris"  # Default ke valid jika error

    @torch.no_grad()
    def extract_embedding(self, image_bytes: bytes) -> list[float]:
        """
        Ekstrak 512-dim embedding dari foto wajah.
        Raise exception jika tidak ada wajah atau lebih dari 1 wajah.
        """
        # Deteksi wajah
        has_face, face_count, boxes = self.detect_faces(image_bytes)
        
        if not has_face:
            raise ValueError("Tidak ada wajah terdeteksi dalam foto")
        
        if face_count > 1:
            raise ValueError(f"Terdeteksi {face_count} wajah. Hanya satu wajah yang diperbolehkan")
        
        # Periksa aksesoris
        is_valid, message = self.check_accessories(image_bytes, boxes)
        if not is_valid:
            raise ValueError(message)

        img = Image.open(io.BytesIO(image_bytes)).convert("RGB")
        tensor = self.transform(img).unsqueeze(0)

        if self.extractor is None:
            raise ValueError("Extractor belum diinisialisasi")

        emb = self.extractor(tensor)
        return emb[0].cpu().numpy().tolist()

    def cosine_similarity(self, emb1: list, emb2: list) -> float:
        a = np.array(emb1, dtype=np.float32)
        b = np.array(emb2, dtype=np.float32)
        a /= (np.linalg.norm(a) + 1e-8)
        b /= (np.linalg.norm(b) + 1e-8)
        return float(np.dot(a, b))

    @torch.no_grad()
    def verify(
        self,
        image_bytes: bytes,
        stored_embedding: list[float],
        threshold: float = None,
    ) -> dict:
        """
        Bandingkan foto wajah vs embedding acuan yang dikirim dari Golang.
        """
        thr = threshold or SIMILARITY_THRESHOLD

        live_emb = self.extract_embedding(image_bytes)
        similarity = self.cosine_similarity(live_emb, stored_embedding)
        matched = similarity >= thr

        # Confidence dari classifier jika tersedia
        confidence = similarity
        if self.classifier and self.class_names:
            img = Image.open(io.BytesIO(image_bytes)).convert("RGB")
            tensor = self.transform(img).unsqueeze(0)
            emb_t = self.extractor(tensor)
            logits = self.classifier(emb_t)
            probs = torch.softmax(logits, dim=1)[0]
            confidence = float(probs.max().item())

        return {
            "matched": matched,
            "similarity": round(similarity, 4),
            "confidence": round(confidence, 4),
            "threshold": thr,
            "message": (
                f"Wajah cocok (similarity={similarity:.1%})"
                if matched else
                f"Wajah tidak cocok (similarity={similarity:.1%}, min={thr:.0%})"
            ),
        }


# Inisialisasi face_svc
face_svc = FaceService()

# =============================================================================
# GEOFENCING
# =============================================================================
def haversine(lat1, lng1, lat2, lng2) -> float:
    R = 6_371_000
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlam = math.radians(lng2 - lng1)
    a = math.sin(dphi/2)**2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlam/2)**2
    return 2 * R * math.atan2(math.sqrt(a), math.sqrt(1-a))


def validate_location(latitude: float, longitude: float,
                      radius_m: float = None) -> dict:
    radius = radius_m or GEOFENCE_RADIUS_M
    distance = haversine(latitude, longitude, OFFICE_LAT, OFFICE_LNG)
    valid = distance <= radius
    return {
        "is_valid": valid,
        "distance_m": round(distance, 1),
        "radius_m": radius,
        "office_lat": OFFICE_LAT,
        "office_lng": OFFICE_LNG,
        "message": (
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
    employee_id: str
    threshold: Optional[float] = None


class GeoRequest(BaseModel):
    latitude: float
    longitude: float
    radius_m: Optional[float] = None


class AttendanceProcessRequest(BaseModel):
    """
    Pipeline lengkap — Golang kirim semua data sekaligus,
    FastAPI kembalikan keputusan akhir.
    """
    employee_id: str
    stored_embedding: list[float]
    latitude: float
    longitude: float
    record_type: str  # "checkin" / "checkout"
    threshold: Optional[float] = None
    radius_m: Optional[float] = None


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
    title="Face Recognition — Hotel Labersa Toba",
    description="""
**Internal untuk face recognition & geofencing.
Dipanggil oleh Golang Backend, bukan langsung oleh client.

### Alur:
1. **Registrasi wajah**: Golang kirim foto → FastAPI ekstrak embedding → 
   kembalikan `embedding[]` → Golang simpan di DB-nya
2. **Absensi**: Golang ambil embedding dari DB → kirim ke FastAPI bersama foto baru → 
   FastAPI bandingkan → kembalikan `matched: true/false`
    """,
    version="1.0.0",
    lifespan=lifespan,
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
        "status": "ok",
        "model_loaded": face_svc.loaded,
        "num_classes": len(face_svc.class_names),
        "office_coords": {"lat": OFFICE_LAT, "lng": OFFICE_LNG},
        "geofence_m": GEOFENCE_RADIUS_M,
        "threshold": SIMILARITY_THRESHOLD,
    }


# ── 1. Ekstrak Embedding (saat registrasi wajah) ──────────────────────────────
@app.post("/face/extract", summary="📸 Ekstrak embedding dari foto wajah")
async def extract_embedding_endpoint(
    photo: UploadFile = File(..., description="Foto wajah (JPG/PNG/WEBP)"),
    employee_id: str = Form(..., description="ID pegawai untuk logging"),
    _=Depends(verify_api_key),
):
    _validate_image_file(photo)
    image_bytes = await photo.read()

    t0 = time.time()

    try:
        # Pastikan service sudah loaded
        if not face_svc.loaded:
            face_svc.load()

        # Panggil method extract_embedding (sudah include validasi multi-face dan aksesoris)
        embedding = face_svc.extract_embedding(image_bytes)
        elapsed = round((time.time() - t0) * 1000, 1)

        logger.info(f"[EXTRACT] employee={employee_id} elapsed={elapsed}ms | embedding dim={len(embedding)}")

        return {
            "success": True,
            "employee_id": employee_id,
            "embedding": embedding,
            "dimension": len(embedding),
            "elapsed_ms": elapsed,
            "message": "Embedding berhasil diekstrak",
        }
    except ValueError as e:
        # Tidak ada wajah, lebih dari 1 wajah, atau aksesoris terdeteksi
        logger.warning(f"[EXTRACT] Validation error: {e}")
        return JSONResponse(
            status_code=400,
            content={
                "success": False,
                "employee_id": employee_id,
                "message": str(e)
            }
        )
    except Exception as e:
        logger.error(f"Error extracting embedding: {e}")
        return JSONResponse(
            status_code=500,
            content={
                "success": False,
                "employee_id": employee_id,
                "message": f"Gagal mengekstrak embedding: {str(e)}"
            }
        )


# ── 2. Verifikasi Wajah (saat absensi) ───────────────────────────────────────
@app.post("/face/verify", summary="🔍 Cocokkan foto vs embedding acuan")
async def verify_face(
    photo: UploadFile = File(..., description="Foto selfie saat absen"),
    data: str = Form(..., description='JSON: {"employee_id":"...","stored_embedding":[...],"threshold":0.75}'),
    _=Depends(verify_api_key),
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

    t0 = time.time()

    try:
        # Pastikan service sudah loaded
        if not face_svc.loaded:
            face_svc.load()

        # Deteksi wajah sebelum verifikasi
        has_face, face_count, boxes = face_svc.detect_faces(image_bytes)
        
        if not has_face:
            elapsed = round((time.time() - t0) * 1000, 1)
            logger.warning(f"[VERIFY] No face detected for employee={req.employee_id}")
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "employee_id": req.employee_id,
                "elapsed_ms": elapsed,
                "message": "Tidak ada wajah terdeteksi dalam foto"
            }
        
        if face_count > 1:
            elapsed = round((time.time() - t0) * 1000, 1)
            logger.warning(f"[VERIFY] Multiple faces ({face_count}) detected for employee={req.employee_id}")
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "employee_id": req.employee_id,
                "elapsed_ms": elapsed,
                "message": f"Terdeteksi {face_count} wajah. Hanya satu wajah yang diperbolehkan"
            }
        
        # Periksa aksesoris
        is_valid, accessory_msg = face_svc.check_accessories(image_bytes, boxes)
        if not is_valid:
            elapsed = round((time.time() - t0) * 1000, 1)
            logger.warning(f"[VERIFY] Accessory detected for employee={req.employee_id}: {accessory_msg}")
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "employee_id": req.employee_id,
                "elapsed_ms": elapsed,
                "message": accessory_msg
            }

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
            "elapsed_ms": elapsed,
        }
    except Exception as e:
        logger.error(f"Error verifying face: {e}")
        elapsed = round((time.time() - t0) * 1000, 1)
        return {
            "matched": False,
            "similarity": 0.0,
            "confidence": 0.0,
            "threshold": req.threshold or SIMILARITY_THRESHOLD,
            "employee_id": req.employee_id,
            "elapsed_ms": elapsed,
            "message": f"Gagal verifikasi: {str(e)}"
        }


# ── 3. Validasi GPS ───────────────────────────────────────────────────────────
@app.post("/geo/validate", summary="📍 Validasi koordinat GPS")
def validate_geo(
    req: GeoRequest,
    _=Depends(verify_api_key),
):
    result = validate_location(req.latitude, req.longitude, req.radius_m)
    logger.info(f"[GEO] lat={req.latitude} lng={req.longitude} valid={result['is_valid']} dist={result['distance_m']}m")
    return result


# ── 4. Pipeline Lengkap — ENDPOINT UTAMA ─────────────────────────────────────
@app.post("/attendance/process", summary="✅ Pipeline lengkap: GPS + Face Verification")
async def process_attendance(
    photo: UploadFile = File(..., description="Foto selfie"),
    data: str = Form(..., description="JSON data"),
    _=Depends(verify_api_key),
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

    # Step 1: Validasi GPS
    geo = validate_location(req.latitude, req.longitude, req.radius_m)

    if not geo["is_valid"]:
        elapsed = round((time.time() - t0) * 1000, 1)
        logger.info(f"[ATTENDANCE] employee={req.employee_id} rejected_gps dist={geo['distance_m']}m")
        return {
            "decision": "rejected_gps",
            "approved": False,
            "employee_id": req.employee_id,
            "record_type": req.record_type,
            "geo": geo,
            "face": None,
            "elapsed_ms": elapsed,
            "message": geo["message"],
        }

    # Step 2: Deteksi wajah
    has_face, face_count, boxes = face_svc.detect_faces(image_bytes)
    
    if not has_face:
        elapsed = round((time.time() - t0) * 1000, 1)
        logger.info(f"[ATTENDANCE] employee={req.employee_id} rejected_face (no face detected)")
        return {
            "decision": "rejected_face",
            "approved": False,
            "employee_id": req.employee_id,
            "record_type": req.record_type,
            "geo": geo,
            "face": {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "message": "Tidak ada wajah terdeteksi dalam foto"
            },
            "elapsed_ms": elapsed,
            "message": "Tidak ada wajah terdeteksi dalam foto",
        }
    
    if face_count > 1:
        elapsed = round((time.time() - t0) * 1000, 1)
        logger.info(f"[ATTENDANCE] employee={req.employee_id} rejected_face ({face_count} faces detected)")
        return {
            "decision": "rejected_face",
            "approved": False,
            "employee_id": req.employee_id,
            "record_type": req.record_type,
            "geo": geo,
            "face": {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "message": f"Terdeteksi {face_count} wajah. Hanya satu wajah yang diperbolehkan"
            },
            "elapsed_ms": elapsed,
            "message": f"Terdeteksi {face_count} wajah",
        }
    
    # Step 3: Periksa aksesoris
    is_valid, accessory_msg = face_svc.check_accessories(image_bytes, boxes)
    if not is_valid:
        elapsed = round((time.time() - t0) * 1000, 1)
        logger.info(f"[ATTENDANCE] employee={req.employee_id} rejected_face (accessory: {accessory_msg})")
        return {
            "decision": "rejected_face",
            "approved": False,
            "employee_id": req.employee_id,
            "record_type": req.record_type,
            "geo": geo,
            "face": {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "message": accessory_msg
            },
            "elapsed_ms": elapsed,
            "message": accessory_msg,
        }

    # Step 4: Verifikasi Wajah
    face = face_svc.verify(image_bytes, req.stored_embedding, req.threshold)

    elapsed = round((time.time() - t0) * 1000, 1)
    approved = face["matched"]
    decision = "approved" if approved else "rejected_face"

    logger.info(
        f"[ATTENDANCE] employee={req.employee_id} type={req.record_type} "
        f"decision={decision} sim={face['similarity']:.3f} dist={geo['distance_m']}m"
    )

    return {
        "decision": decision,
        "approved": approved,
        "employee_id": req.employee_id,
        "record_type": req.record_type,
        "geo": geo,
        "face": face,
        "elapsed_ms": elapsed,
        "message": (
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
    import cv2  # Tambahkan import ini di bagian atas bersama import lainnya
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=int(os.getenv("PORT", "8001")),
        reload=False,
        workers=1,  # 1 worker — model FaceNet tidak thread-safe
    )