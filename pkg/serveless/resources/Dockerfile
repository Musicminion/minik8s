FROM python:3.8-slim

COPY ./  /app

WORKDIR /app

# 换成中国的源
RUN pip3 config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

RUN pip3 install -r ./requirements.txt

EXPOSE 18080

ENTRYPOINT ["python3"]

CMD ["server.py"]

# docker build -t musicminion/func-server:latest .