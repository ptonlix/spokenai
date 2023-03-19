package logo

import (
	"fmt"
	"io"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/ptonlix/spokenai/pkg/color"
)

// see https://patorjk.com/software/taag/#p=testall&f=Graffiti&t=go-gin-api
const ui = `
   _____ _____   ____  _  ________ _   _          _____ 
  / ____|  __ \ / __ \| |/ /  ____| \ | |   /\   |_   _|
 | (___ | |__) | |  | | ' /| |__  |  \| |  /  \    | |  
  \___ \|  ___/| |  | |  < |  __| |     | / /\ \   | |  
  ____) | |    | |__| | . \| |____| |\  |/ ____ \ _| |_ 
 |_____/|_|     \____/|_|\_\______|_| \_/_/    \_\_____|
`

const introduce = `
----------------------------------------------------------------------
Welcome to use 

这是一个基于ChatGPT开发的英语口语AI老师，可以通过该应用与AI老师进
行对话，提升你的英语口语能力。
This is an English speaking AI teacher developed based on ChatGPT, 
through which you can have conversations with AI teachers and improve 
your English speaking skills.

作者/author: Baird
联系作者/Contact the author:
`

func PrintLogo(w io.Writer) {
	fmt.Fprint(w, color.Blue(ui))
}

func PrintIntroduce(w io.Writer) {
	fmt.Fprint(w, color.Yellow(introduce))
}

func PrintQrcode() {
	content := "https://u.wechat.com/EI-mRBDDV9dumugl-s-v06g"
	obj := qrcodeTerminal.New()
	obj.Get(content).Print()
}
