# -*- coding: utf-8 -*-
import pyaudio
import wave
import signal
import schedule
import sys
import getopt
import numpy as np

#最大记录时间 单位s
MAX_RECORD_TIME = 60

Flag = True

def signal_handler(signum, frame):
    global Flag 
    Flag = False

def calnum():
    num = 0
    def inner():
        nonlocal num
        num += 1
        return num
    return inner

def printload(addfunc):
    list = ["\\", "|", "/", "-"]
    timesec = addfunc()
    if timesec >= MAX_RECORD_TIME:
        global Flag
        Flag = False
    print("\r%s 记录%d秒录音" %(list[timesec%4], timesec), end="")    

def Monitor(filepath):
    signal.signal(signal.SIGINT, signal_handler)
    CHUNK = 512
    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    RATE = 48000
    WAVE_OUTPUT_FILENAME = filepath
    p = pyaudio.PyAudio()
    stream = p.open(format=FORMAT,
                    channels=CHANNELS,
                    rate=RATE,
                    input=True,
                    frames_per_buffer=CHUNK)
    print("开始缓存录音")
    frames = []
    while True:
        schedule.every().seconds.do(printload, calnum())
        schedule.run_pending()
        for i in range(0, 100):
            data = stream.read(CHUNK, exception_on_overflow=False)
            frames.append(data)
        audio_data = np.frombuffer(data, dtype=np.short)
        temp = np.max(audio_data)
        if Flag == False :
            print('\n录制结束 总体阈值：',temp) 
            schedule.clear()
            break
    stream.stop_stream()
    stream.close()
    p.terminate()
    wf = wave.open(WAVE_OUTPUT_FILENAME, 'wb')
    wf.setnchannels(CHANNELS)
    wf.setsampwidth(p.get_sample_size(FORMAT))
    wf.setframerate(RATE)
    wf.writeframes(b''.join(frames))
    wf.close()

def main(argv):
    savefile = ''
    try:
        opts, args = getopt.getopt(argv,'hs:',['help', 'savefile='])
        if len(opts) == 0:
            raise Exception("args error")
    except getopt.GetoptError: 
        print('execute example:')
        print('record_audio.py -s <savefile>')
        sys.exit(2)
    except Exception as err:
        print(err, '\nargs: -h -s')
        sys.exit(2)

    for opt, arg in opts:
        if opt in ("-h", "--help"):
            print('execute example:')
            print('record_audio.py -s <savefile>')
            sys.exit()
        elif opt in ("-s", "--savefile"):
            savefile = arg
            
    Monitor(savefile)

if __name__ == "__main__":
   main(sys.argv[1:])