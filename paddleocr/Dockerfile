FROM python:3.10-slim

RUN pip install --no-cache-dir paddlepaddle==2.6.1 paddleocr==2.8.1 -i https://pypi.tuna.tsinghua.edu.cn/simple

EXPOSE 9000

COPY server.py /app/server.py
WORKDIR /app

CMD ["python", "server.py"]
