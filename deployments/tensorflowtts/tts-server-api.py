import io
import numpy as np
import argparse
import soundfile as sf
import yaml

import tensorflow as tf

from tensorflow_tts.inference import TFAutoModel
from tensorflow_tts.inference import AutoProcessor
from flask import Flask, Response, request

# initialize fastspeech2 model.
fastspeech2 = TFAutoModel.from_pretrained("tensorspeech/tts-fastspeech2-ljspeech-en")

# initialize mb_melgan model
mb_melgan = TFAutoModel.from_pretrained("tensorspeech/tts-mb_melgan-ljspeech-en")

# inference
processor = AutoProcessor.from_pretrained("tensorspeech/tts-fastspeech2-ljspeech-en")

app = Flask(__name__)

@app.route('/api/tts', methods=['POST'])
def tts():
    data = request.get_json()
    text = data['text']

    # fastspeech inference
    input_ids = processor.text_to_sequence(text)
    mel_before, mel_after, duration_outputs, _, _ = fastspeech2.inference(
        input_ids=tf.expand_dims(tf.convert_to_tensor(input_ids, dtype=tf.int32), 0),
        speaker_ids=tf.convert_to_tensor([0], dtype=tf.int32),
        speed_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
        f0_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
        energy_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
    )

    # melgan inference
    audio_before = mb_melgan.inference(mel_before)[0, :, 0]
    audio_after = mb_melgan.inference(mel_after)[0, :, 0]

    # save to file

    # Convert audio data to byte stream
    buffer = io.BytesIO()
    sf.write(buffer, audio_after, 22050, format='WAV', subtype='PCM_16')
    audio_bytes = buffer.getvalue()

    # Return audio data as a response with MIME type audio/wav
    return Response(audio_bytes, mimetype='audio/wav')


if __name__ == '__main__':

    parser = argparse.ArgumentParser(description='TensorFlowTTS REST API server')
    parser.add_argument('--host', default='0.0.0.0', help='Hostname to listen on')
    parser.add_argument('--port', type=int, default=5000, help='Port to listen on')
    args = parser.parse_args()

    app.run(host=args.host, port=args.port)