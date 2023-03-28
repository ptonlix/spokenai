## tensorflowtts镜像制作说明

```
curl -X POST -H "Content-Type: application/json" -d '{"text": "Hello, world!"}' -o output.wav -w "%{http_code}" http://localhost:5000/api/tts
```