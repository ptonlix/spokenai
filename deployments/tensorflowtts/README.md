## tensorflowtts镜像说明

```shell
# 自定义构建docker容器
docker build -t tensorflowtts:[tag] .
# 拉取作者已构建的镜像
docker pull ptonlix/tensorflowtts:1.0.9
# 运行镜像
docker run -itd -p 5000:5000 --name spokenai-tts ptonlix/tensorflowtts:1.0.9
```