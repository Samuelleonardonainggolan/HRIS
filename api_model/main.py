import io
import os
import json
import time
import logging
from typing import Optional, Tuple, List, Dict, Any
import cv2
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
MODEL_PATH = os.getenv("MODEL_PATH", r"D:\Semester 6\PA\Training code\output_cpu\facenet_labersa_cpu.pt")
IMAGE_SIZE = int(os.getenv("IMAGE_SIZE", "160"))
SIMILARITY_THRESHOLD = float(os.getenv("SIMILARITY_THRESHOLD", "0.75"))
ANTI_SPOOFING_ENABLED = os.getenv("ANTI_SPOOFING_ENABLED", "1") == "1"
SPOOF_SCORE_THRESHOLD = float(os.getenv("SPOOF_SCORE_THRESHOLD", "0.65"))
FACE_DET_MIN_PROB = float(os.getenv("FACE_DET_MIN_PROB", "0.90"))
LIVENESS_ENABLED = os.getenv("LIVENESS_ENABLED", "1") == "1"
LIVENESS_MIN_FRAMES = int(os.getenv("LIVENESS_MIN_FRAMES", "3"))
LIVENESS_STD_MEAN_THR = float(os.getenv("LIVENESS_STD_MEAN_THR", "0.008"))
LIVENESS_YAW_RANGE_THR = float(os.getenv("LIVENESS_YAW_RANGE_THR", "0.08"))
LIVENESS_MAX_FRAMES = int(os.getenv("LIVENESS_MAX_FRAMES", "6"))
API_KEY = os.getenv("FACE_API_KEY", "labersa-internal-api-key-2026")

# =============================================================================
# MODEL
# =============================================================================
class FaceNetExtractor(nn.Module):
    def __init__(self):
        super().__init__()
        self.backbone = InceptionResnetV1(pretrained="vggface2", classify=False)
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
# FACE SERVICE (stateless) - DENGAN DETEKSI AKSESORIS YANG LEBIH KUAT
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

    def _decode_image(self, image_bytes: bytes) -> np.ndarray:
        """
        Decode bytes gambar ke array RGB uint8 yang eksplisit.
        Menghindari jalur dtype inference yang kadang bermasalah saat pakai PIL langsung.
        """
        image_array = np.frombuffer(image_bytes, dtype=np.uint8)
        image_bgr = cv2.imdecode(image_array, cv2.IMREAD_COLOR)

        if image_bgr is None:
            raise ValueError("File gambar tidak valid atau rusak")

        return cv2.cvtColor(image_bgr, cv2.COLOR_BGR2RGB)

    def load(self):
        try:
            # Inisialisasi MTCNN untuk deteksi wajah yang lebih stabil pada selfie
            self.mtcnn = MTCNN(keep_all=True, device="cpu")
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
                self.mtcnn = MTCNN(keep_all=True, device="cpu")

            img = Image.fromarray(self._decode_image(image_bytes))
            boxes, _ = self.mtcnn.detect(img)

            if boxes is None or len(boxes) == 0:
                return False, 0, []

            valid_boxes = []
            for box in boxes:
                x1, y1, x2, y2 = [float(v) for v in box]
                if (x2 - x1) >= 50 and (y2 - y1) >= 50:
                    valid_boxes.append([x1, y1, x2, y2])

            return len(valid_boxes) > 0, len(valid_boxes), valid_boxes
        except Exception as e:
            logger.error(f"Error detecting faces: {e}")
            return False, 0, []

    def detect_faces_with_landmarks(self, image_bytes: bytes) -> List[Tuple[np.ndarray, np.ndarray, float]]:
        if self.mtcnn is None:
            self.mtcnn = MTCNN(keep_all=True, device="cpu")
        img = Image.fromarray(self._decode_image(image_bytes))
        boxes, probs, landmarks = self.mtcnn.detect(img, landmarks=True)
        if boxes is None or probs is None or landmarks is None:
            return []

        out: List[Tuple[np.ndarray, np.ndarray, float]] = []
        for box, prob, lm in zip(boxes, probs, landmarks):
            if prob is None:
                continue
            p = float(prob)
            if p < FACE_DET_MIN_PROB:
                continue
            x1, y1, x2, y2 = [float(v) for v in box]
            if (x2 - x1) < 50 or (y2 - y1) < 50:
                continue
            out.append((np.array([x1, y1, x2, y2], dtype=np.float32), lm.astype(np.float32), p))
        return out

    def motion_liveness(self, landmarks_seq: List[np.ndarray]) -> Tuple[bool, Dict[str, Any]]:
        if not LIVENESS_ENABLED:
            return True, {"enabled": False}
        if len(landmarks_seq) < LIVENESS_MIN_FRAMES:
            return False, {
                "enabled": True,
                "reason": "not_enough_frames",
                "min_frames": LIVENESS_MIN_FRAMES,
                "frames": len(landmarks_seq),
            }
        feats = []
        yaws = []
        for lm in landmarks_seq:
            le, re, nose, ml, mr = lm
            eye_d = float(np.linalg.norm(le - re) + 1e-9)
            f = np.array([
                np.linalg.norm(nose - le) / eye_d,
                np.linalg.norm(nose - re) / eye_d,
                np.linalg.norm(ml - mr) / eye_d,
                np.linalg.norm(nose - (ml + mr) / 2.0) / eye_d,
            ], dtype=np.float32)
            feats.append(f)
            mid_eye_x = float((le[0] + re[0]) * 0.5)
            yaws.append(float((nose[0] - mid_eye_x) / eye_d))
        feats = np.stack(feats, axis=0)
        std = feats.std(axis=0)
        yaw_range = float(np.max(yaws) - np.min(yaws))
        std_mean = float(std.mean())
        live = bool((std_mean >= LIVENESS_STD_MEAN_THR) and (yaw_range >= LIVENESS_YAW_RANGE_THR))
        return live, {
            "enabled": True,
            "reason": "ok" if live else "liveness_fail",
            "feat_std_mean": std_mean,
            "yaw_range": yaw_range,
            "thr_feat_std_mean": LIVENESS_STD_MEAN_THR,
            "thr_yaw_range": LIVENESS_YAW_RANGE_THR,
            "frames": len(landmarks_seq),
        }

    def _lbp_hist(self, gray_u8: np.ndarray) -> np.ndarray:
        g = gray_u8.astype(np.uint8)
        if g.shape[0] < 3 or g.shape[1] < 3:
            return np.zeros(256, dtype=np.float64)
        c = g[1:-1, 1:-1]
        code = np.zeros_like(c, dtype=np.uint8)
        code |= ((g[:-2, :-2] >= c) << 7).astype(np.uint8)
        code |= ((g[:-2, 1:-1] >= c) << 6).astype(np.uint8)
        code |= ((g[:-2, 2:] >= c) << 5).astype(np.uint8)
        code |= ((g[1:-1, 2:] >= c) << 4).astype(np.uint8)
        code |= ((g[2:, 2:] >= c) << 3).astype(np.uint8)
        code |= ((g[2:, 1:-1] >= c) << 2).astype(np.uint8)
        code |= ((g[2:, :-2] >= c) << 1).astype(np.uint8)
        code |= ((g[1:-1, :-2] >= c) << 0).astype(np.uint8)
        hist = np.bincount(code.reshape(-1), minlength=256).astype(np.float64)
        hist /= (hist.sum() + 1e-12)
        return hist

    def _entropy(self, p: np.ndarray) -> float:
        p = p[p > 0]
        return float(-(p * np.log(p)).sum())

    def _laplacian_var(self, gray_u8: np.ndarray) -> float:
        g = gray_u8.astype(np.float32)
        if g.shape[0] < 3 or g.shape[1] < 3:
            return 0.0
        lap = (-4.0 * g[1:-1, 1:-1] +
               g[:-2, 1:-1] + g[2:, 1:-1] + g[1:-1, :-2] + g[1:-1, 2:])
        return float(lap.var())

    def _fft_peak_ratio(self, gray_u8: np.ndarray) -> float:
        g = gray_u8.astype(np.float32) / 255.0
        g = g - g.mean()
        f = np.fft.fftshift(np.fft.fft2(g))
        mag = np.log1p(np.abs(f))
        h, w = mag.shape
        cy, cx = h // 2, w // 2
        r0 = max(6, int(0.06 * min(h, w)))
        yy, xx = np.ogrid[:h, :w]
        mask = (yy - cy) ** 2 + (xx - cx) ** 2 >= (r0 * r0)
        high = mag[mask]
        if high.size == 0:
            return 0.0
        return float(high.max() / (high.mean() + 1e-9))

    def anti_spoof_check(self, image_bytes: bytes, boxes: list) -> Tuple[bool, str, Dict[str, Any]]:
        if not ANTI_SPOOFING_ENABLED:
            return True, "ok", {"enabled": False}
        if not boxes:
            return False, "Tidak ada wajah terdeteksi (anti-spoofing)", {"enabled": True}

        img = self._decode_image(image_bytes)
        largest_box = max(boxes, key=lambda b: (b[2] - b[0]) * (b[3] - b[1]))
        h, w = img.shape[:2]
        x1, y1, x2, y2 = [int(round(v)) for v in largest_box]
        x1 = max(0, min(w - 1, x1))
        x2 = max(0, min(w, x2))
        y1 = max(0, min(h - 1, y1))
        y2 = max(0, min(h, y2))
        if x2 - x1 < 40 or y2 - y1 < 40:
            return False, "Wajah terlalu kecil untuk verifikasi (anti-spoofing)", {"enabled": True}

        face = img[y1:y2, x1:x2]
        gray = (0.299 * face[..., 0] + 0.587 * face[..., 1] + 0.114 * face[..., 2]).astype(np.uint8)

        lap = self._laplacian_var(gray)
        hist = self._lbp_hist(gray)
        ent = self._entropy(hist)
        peak = self._fft_peak_ratio(gray)

        s_blur = np.clip((18.0 - lap) / 18.0, 0.0, 1.0)
        s_flat = np.clip((4.6 - ent) / 1.2, 0.0, 1.0)
        s_peak = np.clip((peak - 8.0) / 10.0, 0.0, 1.0)
        score = float(np.clip(0.45 * s_blur + 0.35 * s_flat + 0.20 * s_peak, 0.0, 1.0))

        detail = {
            "enabled": True,
            "spoof_score": round(score, 4),
            "lap_var": round(lap, 4),
            "lbp_entropy": round(ent, 4),
            "fft_peak_ratio": round(peak, 4),
            "threshold": SPOOF_SCORE_THRESHOLD,
        }

        if score >= SPOOF_SCORE_THRESHOLD:
            return False, "Foto terindikasi spoofing (foto cetak/layar). Harap ambil selfie langsung dari kamera.", detail
        return True, "ok", detail

    def detect_glasses(self, face_region: np.ndarray) -> Tuple[bool, str]:
        """
        Deteksi kacamata (termasuk kacamata bening) menggunakan beberapa metode
        Disesuaikan untuk kulit putih/polos
        """
        try:
            height, width = face_region.shape[:2]
            
            # Region mata (sekitar 1/3 atas wajah)
            eye_region_y1 = int(height * 0.2)
            eye_region_y2 = int(height * 0.5)
            eye_region = face_region[eye_region_y1:eye_region_y2, :]
            
            if eye_region.size == 0:
                return False, "Region mata tidak valid"
            
            # 1. Deteksi tepi menggunakan Canny (frame kacamata)
            gray_eye = cv2.cvtColor(eye_region, cv2.COLOR_RGB2GRAY)
            edges = cv2.Canny(gray_eye, 50, 150)
            
            # Hitung rasio tepi (edge density)
            edge_ratio = np.sum(edges > 0) / edges.size if edges.size > 0 else 0
            
            # 2. Deteksi garis horizontal yang kuat (frame kacamata)
            horizontal_lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=50, minLineLength=30, maxLineGap=10)
            
            # 3. Analisis tekstur untuk kacamata bening - PERBAIKAN UTAMA
            from skimage.feature import local_binary_pattern
            
            # Gunakan radius lebih besar untuk menangkap pola lebih luas
            lbp = local_binary_pattern(gray_eye, 16, 2, method='uniform')
            lbp_hist, _ = np.histogram(lbp.ravel(), bins=np.arange(0, 19), range=(0, 18))
            lbp_hist = lbp_hist / (lbp_hist.sum() + 1e-8)
            
            # Hitung entropy (ketidak-teraturan) - alternatif dari uniformity
            entropy = -np.sum(lbp_hist * np.log2(lbp_hist + 1e-8))
            
            # 4. Deteksi refleksi/silau pada kacamata
            brightness_std = np.std(gray_eye)
            
            # ====================================================
            # PARAMETER YANG DIOPTIMASI UNTUK KULIT PUTIH
            # ====================================================
            
            # Texture uniformity untuk kulit putih alami lebih rendah
            # Tapi kacamata bening menyebabkan pola yang SANGAT berbeda
            # Kita gunakan ENTROPY sebagai pengganti uniformity
            
            # Batas entropy untuk kulit normal (lebih tinggi = lebih acak)
            # Kulit putih: entropy sekitar 2.5 - 3.5
            # Kacamata bening: entropy bisa turun drastis (< 2.0) karena pola teratur
            ENTROPY_THRESHOLD = 2.2  # Jika entropy < 2.2, curiga kacamata
            
            # Edge ratio untuk frame kacamata
            EDGE_RATIO_THRESHOLD = 0.28  # Dinaikkan untuk kulit putih
            
            # Brightness std untuk refleksi
            BRIGHTNESS_STD_THRESHOLD = 80  # Dinaikkan
            
            # ====================================================
            # VALIDASI TAMBAHAN - CEK KULIT ASLI
            # ====================================================
            
            # Ambil sampel kulit dari pipi (area bawah mata)
            cheek_y1 = int(height * 0.5)
            cheek_y2 = int(height * 0.7)
            cheek_region = face_region[cheek_y1:cheek_y2, :]
            
            if cheek_region.size > 0:
                gray_cheek = cv2.cvtColor(cheek_region, cv2.COLOR_RGB2GRAY)
                cheek_std = np.std(gray_cheek)
                cheek_mean = np.mean(gray_cheek)
                
                # Kulit asli memiliki variasi yang konsisten
                # Jika area mata dan pipi memiliki karakteristik mirip, mungkin bukan kacamata
                skin_similarity = abs(brightness_std - cheek_std) / (cheek_std + 1e-8)
            else:
                skin_similarity = 1.0
            
            # ====================================================
            # LOGIKA KEPUTUSAN UTAMA - FOKUS PADA ENTROPY
            # ====================================================
            
            # LOG 1: Cek entropy (indikator utama kacamata bening)
            if entropy < ENTROPY_THRESHOLD:
                # Tapi cek juga skin similarity untuk validasi
                if skin_similarity > 0.3:  # Area mata berbeda dari pipi
                    return True, f"Terdeteksi pola tidak wajar di area mata (entropy={entropy:.2f} < {ENTROPY_THRESHOLD})"
                else:
                    # Jika mirip dengan kulit pipi, mungkin false positive
                    logger.info(f"[GLASSES] False positive terdeteksi: entropy={entropy:.2f} tapi skin_similarity={skin_similarity:.2f}")
                    # Lanjutkan ke pengecekan lain
            
            # LOG 2: Deteksi frame kacamata (edge dan garis)
            if edge_ratio > EDGE_RATIO_THRESHOLD:
                return True, f"Terdeteksi bingkai kacamata (edge={edge_ratio:.2f})"
            
            if horizontal_lines is not None and len(horizontal_lines) > 4:
                return True, "Terdeteksi garis horizontal (frame kacamata)"
            
            # LOG 3: Deteksi refleksi dengan validasi tambahan
            if brightness_std > BRIGHTNESS_STD_THRESHOLD:
                # Cek apakah ini memang refleksi atau hanya kulit cerah
                # Kulit cerah punya std tinggi tapi juga mean tinggi
                if brightness_std > 90 and skin_similarity < 0.2:
                    return True, f"Terdeteksi refleksi cahaya (std={brightness_std:.1f})"
            
            # Log untuk debugging
            logger.info(f"[GLASSES] edge={edge_ratio:.3f}, entropy={entropy:.3f}, brightness_std={brightness_std:.1f}, skin_sim={skin_similarity:.2f}")
            
            return False, "Tidak terdeteksi kacamata"
            
        except Exception as e:
            logger.error(f"Error detecting glasses: {e}")
            return False, "Error deteksi kacamata"

    def detect_mask(self, face_region: np.ndarray) -> Tuple[bool, str]:
        """
        Deteksi masker menggunakan analisis warna dan tekstur
        Dioptimalkan untuk kulit putih/polos
        """
        try:
            height, width = face_region.shape[:2]
            
            # Region mulut dan dagu (sepertiga bawah wajah)
            mouth_region_y1 = int(height * 0.6)
            mouth_region_y2 = height
            mouth_region = face_region[mouth_region_y1:mouth_region_y2, :]
            
            if mouth_region.size == 0:
                return False, "Region mulut tidak valid"
            
            # ====================================================
            # AMBIL SAMPEL KULIT ASLI (dari pipi/dahi)
            # ====================================================
            # Ambil region kulit dari bagian atas wajah (pipi/dahi)
            skin_region_y1 = int(height * 0.3)
            skin_region_y2 = int(height * 0.5)
            skin_region = face_region[skin_region_y1:skin_region_y2, int(width*0.3):int(width*0.7)]
            
            # ====================================================
            # ANALISIS WARNA
            # ====================================================
            # Konversi ke HSV untuk analisis warna
            hsv_mouth = cv2.cvtColor(mouth_region, cv2.COLOR_RGB2HSV)
            hsv_skin = cv2.cvtColor(skin_region, cv2.COLOR_RGB2HSV) if skin_region.size > 0 else None
            
            # Hitung statistik warna area mulut
            mouth_hue_mean = np.mean(hsv_mouth[:,:,0])
            mouth_sat_mean = np.mean(hsv_mouth[:,:,1])
            mouth_val_mean = np.mean(hsv_mouth[:,:,2])
            
            # Statistik kulit sebagai pembanding
            if hsv_skin is not None and hsv_skin.size > 0:
                skin_hue_mean = np.mean(hsv_skin[:,:,0])
                skin_sat_mean = np.mean(hsv_skin[:,:,1])
                skin_val_mean = np.mean(hsv_skin[:,:,2])
                
                # Hitung perbedaan warna antara mulut dan kulit
                hue_diff = abs(mouth_hue_mean - skin_hue_mean)
                sat_diff = abs(mouth_sat_mean - skin_sat_mean)
                val_diff = abs(mouth_val_mean - skin_val_mean)
            else:
                hue_diff = 30  # Default tinggi
                sat_diff = 50
                val_diff = 50
            
            # ====================================================
            # DETEKSI WARNA MASKER UMUM
            # ====================================================
            total_pixels = mouth_region.shape[0] * mouth_region.shape[1]
            
            # Masker biru (medis)
            blue_mask = np.sum((hsv_mouth[:,:,0] > 90) & (hsv_mouth[:,:,0] < 130) & 
                            (hsv_mouth[:,:,1] > 70) & (hsv_mouth[:,:,2] > 50))
            blue_ratio = blue_mask / total_pixels
            
            # Masker hijau (medis)
            green_mask = np.sum((hsv_mouth[:,:,0] > 35) & (hsv_mouth[:,:,0] < 85) & 
                            (hsv_mouth[:,:,1] > 60) & (hsv_mouth[:,:,2] > 50))
            green_ratio = green_mask / total_pixels
            
            # Masker hitam
            black_mask = np.sum((hsv_mouth[:,:,2] < 60) & (hsv_mouth[:,:,1] < 50))
            black_ratio = black_mask / total_pixels
            
            # Masker putih (sering digunakan)
            white_mask = np.sum((hsv_mouth[:,:,2] > 180) & (hsv_mouth[:,:,1] < 40))
            white_ratio = white_mask / total_pixels
            
            # ====================================================
            # DETEKSI TEKSTUR
            # ====================================================
            gray_mouth = cv2.cvtColor(mouth_region, cv2.COLOR_RGB2GRAY)
            
            # Edge density (detail area mulut)
            edges = cv2.Canny(gray_mouth, 50, 150)
            edge_density = np.sum(edges > 0) / total_pixels if total_pixels > 0 else 0
            
            # Texture analysis dengan LBP
            from skimage.feature import local_binary_pattern
            lbp = local_binary_pattern(gray_mouth, 8, 1, method='uniform')
            lbp_hist, _ = np.histogram(lbp.ravel(), bins=np.arange(0, 11), range=(0, 10))
            lbp_hist = lbp_hist / (lbp_hist.sum() + 1e-8)
            texture_uniformity = np.sum(lbp_hist ** 2)
            
            # ====================================================
            # VALIDASI BIBIR
            # ====================================================
            # Deteksi warna bibir (merah) di area mulut
            # Konversi ke ruang warna yang lebih baik untuk deteksi bibir
            lab_mouth = cv2.cvtColor(mouth_region, cv2.COLOR_RGB2LAB)
            
            # Bibir memiliki nilai 'a' yang tinggi (merah-hijau)
            a_channel = lab_mouth[:,:,1]
            lip_mask = a_channel > np.percentile(a_channel, 70)  # Top 30% nilai 'a'
            lip_ratio = np.sum(lip_mask) / total_pixels if total_pixels > 0 else 0
            
            # ====================================================
            # PARAMETER THRESHOLD UNTUK KULIT PUTIH
            # ====================================================
            
            # Threshold untuk warna masker
            BLUE_THRESHOLD = 0.25      # Turunkan dari 0.3
            GREEN_THRESHOLD = 0.25     # Turunkan dari 0.3
            BLACK_THRESHOLD = 0.40     # Turunkan dari 0.5
            WHITE_THRESHOLD = 0.35     # Threshold baru untuk masker putih
            
            # Threshold untuk perbedaan warna dengan kulit
            HUE_DIFF_THRESHOLD = 20    # Perbedaan hue > 20
            SAT_DIFF_THRESHOLD = 40    # Perbedaan saturasi > 40
            
            # Threshold untuk tekstur
            EDGE_DENSITY_THRESHOLD = 0.03      # Turunkan dari 0.05
            TEXTURE_UNIFORMITY_THRESHOLD = 0.25 # Texture uniformity untuk kulit normal
            
            # Threshold untuk deteksi bibir
            LIP_RATIO_THRESHOLD = 0.15  # Minimal 15% area adalah bibir
            
            # ====================================================
            # LOGIKA KEPUTUSAN
            # ====================================================
            
            # LOG 1: Cek apakah ada bibir (indikator tidak pakai masker)
            if lip_ratio > LIP_RATIO_THRESHOLD:
                # Ada bibir terdeteksi, kemungkinan tidak pakai masker
                # Tapi tetap cek warna masker jika sangat dominan
                if blue_ratio > BLUE_THRESHOLD or green_ratio > GREEN_THRESHOLD or black_ratio > BLACK_THRESHOLD or white_ratio > WHITE_THRESHOLD:
                    # Warna masker dominan meski ada bibir? Curiga
                    # Tapi perlu validasi lebih lanjut
                    if hue_diff > HUE_DIFF_THRESHOLD and sat_diff > SAT_DIFF_THRESHOLD:
                        return True, f"Terdeteksi warna tidak wajar di area mulut"
                
                # Ada bibir, kemungkinan tidak pakai masker
                logger.info(f"[MASK] Lip detected: {lip_ratio:.2f}")
                return False, "Bibir terdeteksi"
            
            # LOG 2: Deteksi warna masker spesifik
            if blue_ratio > BLUE_THRESHOLD:
                return True, f"Terdeteksi warna biru dominan (masker)"
            
            if green_ratio > GREEN_THRESHOLD:
                return True, f"Terdeteksi warna hijau dominan (masker medis)"
            
            if black_ratio > BLACK_THRESHOLD:
                return True, f"Terdeteksi area gelap dominan (masker hitam)"
            
            if white_ratio > WHITE_THRESHOLD:
                return True, f"Terdeteksi warna putih dominan (masker)"
            
            # LOG 3: Deteksi berdasarkan perbedaan warna dengan kulit
            if hue_diff > HUE_DIFF_THRESHOLD and sat_diff > SAT_DIFF_THRESHOLD:
                # Area mulut sangat berbeda warna dari kulit
                # Tapi pastikan bukan karena bibir
                if lip_ratio < 0.1:  # Bibir tidak terdeteksi
                    return True, f"Warna area mulut berbeda dari kulit (kemungkinan masker)"
            
            # LOG 4: Deteksi berdasarkan edge density (detail)
            if edge_density < EDGE_DENSITY_THRESHOLD:
                # Kurang detail, tapi pastikan bukan karena kulit halus
                if texture_uniformity < TEXTURE_UNIFORMITY_THRESHOLD:
                    return True, f"Area mulut terlalu halus (kemungkinan masker)"
            
            # Log untuk debugging
            logger.info(f"[MASK] blue={blue_ratio:.3f}, green={green_ratio:.3f}, black={black_ratio:.3f}, white={white_ratio:.3f}, lip={lip_ratio:.3f}, edge={edge_density:.3f}")
            
            return False, "Tidak terdeteksi masker"
            
        except Exception as e:
            logger.error(f"Error detecting mask: {e}")
            return False, "Error deteksi masker"

    def detect_hat(self, face_region: np.ndarray, full_image: np.ndarray, box: list) -> Tuple[bool, str]:
        """
        Deteksi topi/aksesoris kepala dengan menganalisis area di atas wajah.
        PERBAIKAN: Threshold diperketat agar rambut alami tidak salah terdeteksi sebagai topi.
        """
        try:
            x1, y1, x2, y2 = [int(b) for b in box]

            # Area di atas wajah — ambil hanya 40% di atas bounding box wajah
            head_top = max(0, y1 - int((y2 - y1) * 0.4))
            head_region = full_image[head_top:y1, x1:x2]

            if head_region.size == 0:
                return False, "Region kepala tidak valid"

            total_pixels = head_region.shape[0] * head_region.shape[1]
            if total_pixels < 100:
                return False, "Region terlalu kecil"

            gray_head = cv2.cvtColor(head_region, cv2.COLOR_RGB2GRAY)
            hsv_head  = cv2.cvtColor(head_region, cv2.COLOR_RGB2HSV)

            # ── 1. Edge density ─────────────────────────────────────────────
            # Rambut memiliki tekstur tinggi → edge density bisa mencapai 0.3.
            # Topi keras/baseball memiliki garis-garis SANGAT jelas dan tepi lurus.
            # Naikkan threshold dari 0.3 → 0.65 untuk menghindari false-positive rambut.
            edges = cv2.Canny(gray_head, 80, 200)
            edge_density = np.sum(edges > 0) / edges.size

            # Tambahan: pastikan ada garis horizontal panjang (ciri topi, bukan rambut)
            lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=60,
                                    minLineLength=int((x2-x1)*0.5),
                                    maxLineGap=10)
            has_strong_horizontal = False
            if lines is not None:
                for line in lines:
                    x_a, y_a, x_b, y_b = line[0]
                    angle = abs(np.degrees(np.arctan2(y_b - y_a, x_b - x_a)))
                    if angle < 10 or angle > 170:  # hampir horizontal
                        has_strong_horizontal = True
                        break

            # Hanya flag jika edge density SANGAT tinggi DAN ada garis horizontal
            if edge_density > 0.65 and has_strong_horizontal:
                return True, "Terdeteksi tepi tajam di kepala (kemungkinan topi)"

            # ── 2. Warna area kepala ─────────────────────────────────────────
            hue   = hsv_head[:, :, 0]
            sat   = hsv_head[:, :, 1]
            val   = hsv_head[:, :, 2]

            # Warna rambut alami: hitam (val rendah), coklat (hue ~10-30, sat sedang),
            # pirang (hue ~20-35, val tinggi), putih/abu (sat rendah).
            # Kita KECUALIKAN warna-warna rambut tersebut sebelum menghitung warna topi.

            # Mask piksel yang BUKAN rambut alami (saturasi tinggi + hue di luar rambut)
            # Rambut alami umumnya: saturation < 150 ATAU hue dalam rentang rambut
            is_natural_hair = (
                (sat < 60) |                                           # hitam/abu/putih
                ((hue >= 5)  & (hue <= 40) & (sat < 180)) |           # coklat/pirang
                (val < 50)                                             # sangat gelap
            )

            # Piksel non-rambut (kandidat warna aksesoris)
            non_hair_mask = ~is_natural_hair
            non_hair_ratio = np.sum(non_hair_mask) / total_pixels

            # Jika mayoritas piksel terlihat sebagai rambut alami, langsung pass
            if non_hair_ratio < 0.30:
                logger.info(f"[HAT] Kemungkinan rambut alami (non_hair={non_hair_ratio:.2f}), skip")
                return False, "Tidak terdeteksi topi"

            # Hanya cek warna mencolok pada piksel NON-rambut
            non_hair_hue = hue[non_hair_mask]
            non_hair_total = len(non_hair_hue)

            if non_hair_total == 0:
                return False, "Tidak terdeteksi topi"

            # Topi biru solid (hue 100-130, saturasi tinggi)
            blue_topi  = np.sum(
                non_hair_mask &
                (hue > 100) & (hue < 130) &
                (sat > 100)
            ) / total_pixels

            # Topi merah solid (hue <10 atau >160, saturasi tinggi)
            red_topi   = np.sum(
                non_hair_mask &
                ((hue < 10) | (hue > 160)) &
                (sat > 100)
            ) / total_pixels

            # Topi hijau/kuning mencolok
            green_topi = np.sum(
                non_hair_mask &
                (hue > 35) & (hue < 85) &
                (sat > 120)
            ) / total_pixels

            # ── Threshold diperketat: 0.4 → 0.55 ───────────────────────────
            logger.info(
                f"[HAT] edge={edge_density:.3f} non_hair={non_hair_ratio:.2f} "
                f"blue={blue_topi:.3f} red={red_topi:.3f} green={green_topi:.3f}"
            )

            if blue_topi > 0.55:
                return True, "Terdeteksi warna biru mencolok di kepala (kemungkinan topi)"

            if red_topi > 0.55:
                return True, "Terdeteksi warna merah mencolok di kepala (kemungkinan topi)"

            if green_topi > 0.55:
                return True, "Terdeteksi warna mencolok di kepala (kemungkinan topi)"

            return False, "Tidak terdeteksi topi"

        except Exception as e:
            logger.error(f"Error detecting hat: {e}")
            return False, "Error deteksi topi"

    def check_accessories(self, image_bytes: bytes, boxes: list) -> Tuple[bool, str]:
        """
        Periksa apakah ada aksesoris yang menutupi wajah
        Returns: (is_valid, message)
        """
        try:
            img_array = self._decode_image(image_bytes)
            
            if len(boxes) == 0:
                return False, "Tidak ada wajah terdeteksi"
            
            # Ambil wajah pertama (yang terbesar) untuk diperiksa
            largest_box = max(boxes, key=lambda b: (b[2]-b[0]) * (b[3]-b[1]))
            x1, y1, x2, y2 = [int(b) for b in largest_box]
            
            # Pastikan koordinat dalam batas gambar
            x1 = max(0, x1)
            y1 = max(0, y1)
            x2 = min(img_array.shape[1], x2)
            y2 = min(img_array.shape[0], y2)
            
            face_region = img_array[y1:y2, x1:x2]
            
            if face_region.size == 0:
                return False, "Region wajah tidak valid"
            
            # 1. Deteksi kacamata (termasuk bening)
            has_glasses, glasses_msg = self.detect_glasses(face_region)
            if has_glasses:
                return False, f"{glasses_msg}. Harap lepas kacamata."
            
            # 2. Deteksi masker
            has_mask, mask_msg = self.detect_mask(face_region)
            if has_mask:
                return False, f"{mask_msg}. Harap lepas masker."
            
            # 3. Deteksi topi
            has_hat, hat_msg = self.detect_hat(face_region, img_array, largest_box)
            if has_hat:
                return False, f"{hat_msg}. Harap lepas topi/aksesoris kepala."
            
            return True, "Wajah valid, tidak ada aksesoris terdeteksi"
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

        ok, msg, _ = self.anti_spoof_check(image_bytes, boxes)
        if not ok:
            raise ValueError(msg)
        
        # Periksa aksesoris
        is_valid, message = self.check_accessories(image_bytes, boxes)
        if not is_valid:
            raise ValueError(message)

        full_image = self._decode_image(image_bytes)
        largest_box = max(boxes, key=lambda b: (b[2]-b[0]) * (b[3]-b[1]))
        x1, y1, x2, y2 = [max(0, int(v)) for v in largest_box]
        face_crop = full_image[y1:y2, x1:x2]
        if face_crop.size == 0:
            raise ValueError("Region wajah tidak valid")

        img = Image.fromarray(face_crop)
        tensor = self.transform(img).unsqueeze(0)

        if self.extractor is None:
            raise ValueError("Extractor belum diinisialisasi")

        emb = self.extractor(tensor)
        return emb[0].detach().cpu().tolist()

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

        try:
            live_emb = self.extract_embedding(image_bytes)
        except ValueError as e:
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": thr,
                "message": str(e),
            }
        similarity = self.cosine_similarity(live_emb, stored_embedding)
        matched = similarity >= thr

        # Confidence dari classifier jika tersedia
        confidence = similarity
        if self.classifier and self.class_names:
            full_image = self._decode_image(image_bytes)
            has_face, _, boxes = self.detect_faces(image_bytes)
            if has_face and boxes:
                largest_box = max(boxes, key=lambda b: (b[2]-b[0]) * (b[3]-b[1]))
                x1, y1, x2, y2 = [max(0, int(v)) for v in largest_box]
                face_crop = full_image[y1:y2, x1:x2]
                img = Image.fromarray(face_crop if face_crop.size > 0 else full_image)
            else:
                img = Image.fromarray(full_image)
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
**Internal untuk face recognition.
Dipanggil oleh Golang Backend, bukan langsung oleh client.

### Alur:
1. **Registrasi wajah**: Golang kirim foto → FastAPI ekstrak embedding → 
   kembalikan `embedding[]` → Golang simpan di DB-nya
2. **Verifikasi**: Golang ambil embedding dari DB → kirim ke FastAPI bersama foto baru → 
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

        ok, msg, detail = face_svc.anti_spoof_check(image_bytes, boxes)
        if not ok:
            elapsed = round((time.time() - t0) * 1000, 1)
            logger.warning(f"[VERIFY] Anti-spoof rejected employee={req.employee_id} detail={detail}")
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "employee_id": req.employee_id,
                "elapsed_ms": elapsed,
                "message": msg,
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

@app.post("/face/verify_burst", summary="🔍 Cocokkan multi-foto (burst) + liveness gerak")
async def verify_face_burst(
    photos: List[UploadFile] = File(..., description="3-6 foto selfie berurutan (burst)"),
    data: str = Form(..., description='JSON: {"employee_id":"...","stored_embedding":[...],"threshold":0.75}'),
    _=Depends(verify_api_key),
):
    try:
        req = VerifyRequest(**json.loads(data))
    except Exception as e:
        raise HTTPException(400, f"Format 'data' tidak valid: {e}")

    if len(req.stored_embedding) != 512:
        raise HTTPException(400, f"stored_embedding harus 512 dimensi, dapat {len(req.stored_embedding)}")

    if not photos:
        raise HTTPException(400, "photos wajib diisi")
    if len(photos) < LIVENESS_MIN_FRAMES:
        raise HTTPException(400, f"Minimal {LIVENESS_MIN_FRAMES} foto untuk burst")
    if len(photos) > LIVENESS_MAX_FRAMES:
        raise HTTPException(400, f"Maksimal {LIVENESS_MAX_FRAMES} foto untuk burst")

    for p in photos:
        _validate_image_file(p)

    t0 = time.time()

    try:
        if not face_svc.loaded:
            face_svc.load()

        frames = [await p.read() for p in photos]
        landmarks_seq: List[np.ndarray] = []
        frame_details = []

        best_idx = 0
        best_quality = -1.0

        for i, frame_bytes in enumerate(frames):
            dets = face_svc.detect_faces_with_landmarks(frame_bytes)
            if len(dets) == 0:
                elapsed = round((time.time() - t0) * 1000, 1)
                return {
                    "matched": False,
                    "similarity": 0.0,
                    "confidence": 0.0,
                    "threshold": req.threshold or SIMILARITY_THRESHOLD,
                    "employee_id": req.employee_id,
                    "elapsed_ms": elapsed,
                    "message": f"Frame #{i+1}: tidak ada wajah terdeteksi",
                }
            if len(dets) > 1:
                elapsed = round((time.time() - t0) * 1000, 1)
                return {
                    "matched": False,
                    "similarity": 0.0,
                    "confidence": 0.0,
                    "threshold": req.threshold or SIMILARITY_THRESHOLD,
                    "employee_id": req.employee_id,
                    "elapsed_ms": elapsed,
                    "message": f"Frame #{i+1}: terdeteksi lebih dari satu wajah",
                }

            box, lm, prob = dets[0]
            box_list = [[float(box[0]), float(box[1]), float(box[2]), float(box[3])]]

            is_valid, accessory_msg = face_svc.check_accessories(frame_bytes, box_list)
            if not is_valid:
                elapsed = round((time.time() - t0) * 1000, 1)
                return {
                    "matched": False,
                    "similarity": 0.0,
                    "confidence": 0.0,
                    "threshold": req.threshold or SIMILARITY_THRESHOLD,
                    "employee_id": req.employee_id,
                    "elapsed_ms": elapsed,
                    "message": f"Frame #{i+1}: {accessory_msg}",
                }

            ok, msg, detail = face_svc.anti_spoof_check(frame_bytes, box_list)
            if not ok:
                elapsed = round((time.time() - t0) * 1000, 1)
                return {
                    "matched": False,
                    "similarity": 0.0,
                    "confidence": 0.0,
                    "threshold": req.threshold or SIMILARITY_THRESHOLD,
                    "employee_id": req.employee_id,
                    "elapsed_ms": elapsed,
                    "message": f"Frame #{i+1}: {msg}",
                    "anti_spoof": detail,
                }

            img = face_svc._decode_image(frame_bytes)
            h, w = img.shape[:2]
            x1, y1, x2, y2 = [int(round(v)) for v in box.tolist()]
            x1 = max(0, min(w - 1, x1))
            x2 = max(0, min(w, x2))
            y1 = max(0, min(h - 1, y1))
            y2 = max(0, min(h, y2))
            face = img[y1:y2, x1:x2]
            if face.size == 0:
                quality = 0.0
            else:
                gray = (0.299 * face[..., 0] + 0.587 * face[..., 1] + 0.114 * face[..., 2]).astype(np.uint8)
                quality = face_svc._laplacian_var(gray)
            if quality > best_quality:
                best_quality = float(quality)
                best_idx = i

            landmarks_seq.append(lm)
            frame_details.append({
                "frame": i + 1,
                "face_prob": round(float(prob), 4),
                "quality_lap_var": round(float(quality), 4),
                "anti_spoof": detail,
            })

        live, live_detail = face_svc.motion_liveness(landmarks_seq)
        if not live:
            elapsed = round((time.time() - t0) * 1000, 1)
            return {
                "matched": False,
                "similarity": 0.0,
                "confidence": 0.0,
                "threshold": req.threshold or SIMILARITY_THRESHOLD,
                "employee_id": req.employee_id,
                "elapsed_ms": elapsed,
                "message": "Liveness gagal. Ambil burst sambil gerakkan kepala sesuai instruksi (kiri-kanan).",
                "liveness": live_detail,
                "frames": frame_details,
            }

        result = face_svc.verify(frames[best_idx], req.stored_embedding, req.threshold)
        elapsed = round((time.time() - t0) * 1000, 1)
        return {
            **result,
            "employee_id": req.employee_id,
            "elapsed_ms": elapsed,
            "burst_best_frame": best_idx + 1,
            "liveness": live_detail,
            "frames": frame_details,
        }
    except Exception as e:
        logger.error(f"Error verifying face burst: {e}")
        elapsed = round((time.time() - t0) * 1000, 1)
        return {
            "matched": False,
            "similarity": 0.0,
            "confidence": 0.0,
            "threshold": req.threshold or SIMILARITY_THRESHOLD,
            "employee_id": req.employee_id,
            "elapsed_ms": elapsed,
            "message": f"Gagal verifikasi burst: {str(e)}",
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
        host="0.0.0.0",
        port=int(os.getenv("PORT", "8001")),
        reload=False,
        workers=1,
    )
