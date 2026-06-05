FROM python:3.9

# Install dependencies sistem yang mungkin dibutuhkan oleh OpenCV/Pillow (karena menggunakan FaceNet/Torch)
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Buat user non-root seperti yang diwajibkan Hugging Face
RUN useradd -m -u 1000 user
USER user
ENV PATH="/home/user/.local/bin:$PATH"

WORKDIR /app

# Copy requirement dan install
COPY --chown=user ./requirements.txt requirements.txt
RUN pip install --no-cache-dir --upgrade -r requirements.txt

# Copy seluruh source code
COPY --chown=user . /app

# Jalankan uvicorn di port 7860 sesuai permintaan Hugging Face
# File utama Anda adalah main.py, jadi kita gunakan main:app (bukan app:app)
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "7860"]
