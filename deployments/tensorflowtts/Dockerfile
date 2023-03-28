FROM tensorflow/tensorflow:2.6.0

# 安装必要的依赖
RUN apt-get update  &&\
    apt-get install -y libsndfile1

# 安装TensorFlowTTS
RUN pip install -i https://mirrors.aliyun.com/pypi/simple/ TensorFlowTTS flask

# 安装
RUN apt-get install -y git
RUN pip install git+https://github.com/repodiac/german_transliterate.git#egg=german_transliterate
RUN pip install --upgrade numpy numba

# 安装 
ADD tts-server-api.py /app/tts-server-api.py

# 运行REST API服务器
CMD python /app/tts-server-api.py --host 0.0.0.0 --port 5000
