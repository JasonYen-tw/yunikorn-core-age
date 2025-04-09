# 使用輕量級基底映像檔，例如 alpine
FROM alpine:latest

# 如有需要安裝依賴，可以在此安裝（例如 ca-certificates）
# RUN apk add --no-cache ca-certificates

# 複製編譯後的二進位檔到映像檔內
COPY build/simplescheduler /usr/local/bin/yunikorn-core

# 賦予執行權限（若需要）
RUN chmod +x /usr/local/bin/yunikorn-core

# 設定容器啟動時執行的命令
ENTRYPOINT ["/usr/local/bin/yunikorn-core"]

