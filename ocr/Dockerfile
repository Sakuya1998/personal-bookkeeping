FROM python:3.11-slim

RUN pip install --no-cache-dir \
    gunthercox-ocr-service \
    "paddlepaddle==2.6.1" \
    "paddleocr==2.7.3"

ENV PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK=True

EXPOSE 5000

CMD ["gunicorn", "app.api:app", "--bind", "0.0.0.0:5000", "--workers", "4"]
