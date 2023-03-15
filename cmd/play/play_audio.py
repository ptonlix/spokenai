# -*- coding: utf-8 -*-
import pyaudio
import wave
import sys
import getopt

def Play(filepath):	
    audio = pyaudio.PyAudio()  # 新建一个PyAudio对象
    # wave.open跟python内置的open有点类似，从wf中可以获取音频基本信息
    with wave.open(filepath, "rb") as wf:
        stream = audio.open(format=audio.get_format_from_width(wf.getsampwidth()),  # 指定数据类型是int16，也就是一个数据点占2个字节；paInt16=8，paInt32=2，不明其含义，先不管
                        channels=wf.getnchannels(),  # 声道数，1或2
                        rate=wf.getframerate(),  # 采样率，44100或16000居多
                        frames_per_buffer=1024,  # 每个块包含多少帧，详见下文解释
                        output=True)  # 表示这是一个输出流，要对外播放的
        # getnframes获取整个文件的帧数，readframes读取n帧，两者结合就是读取整个文件所有帧
        stream.write(wf.readframes(wf.getnframes()))  # 把数据写进流对象
        stream.stop_stream()  # stop后在start_stream()之前不可再read或write
        stream.close()  # 关闭这个流对象
        audio.terminate()  # 关闭PyAudio对象

def main(argv):
    filepath = ''
    try:
        opts, args = getopt.getopt(argv,'hi:',['help', 'filepath='])
        if len(opts) == 0:
            raise Exception("args error")
    except getopt.GetoptError: 
        print('execute example:')
        print('play_audio.py -i <filepath>')
        sys.exit(2)
    except Exception as err:
        print(err, '\nargs: -h -i')
        sys.exit(2)

    for opt, arg in opts:
        if opt in ("-h", "--help"):
            print('execute example:')
            print('record_audio.py -i <filepath>')
            sys.exit()
        elif opt in ("-i", "--filepath"):
            savefile = arg
            
    Play(savefile)

if __name__ == "__main__":
   main(sys.argv[1:])