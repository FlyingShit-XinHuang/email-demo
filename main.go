package main

import (
	"gopkg.in/gomail.v2"
	"log"
	"time"
	//"encoding/base64"
	"encoding/base64"
	"fmt"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	//mail2 "net/mail"
	//"github.com/emersion/go-message/charset"
	"io"
	"io/ioutil"
)

const user = "foo@hello.world"
const password = "xxxxxxxx"
var sended map[string]string

func main() {
	ch := make(chan *gomail.Message)
	startSender(ch)
	msg := gomail.NewMessage()
	mid, appid := "test", "test"

	msg.SetHeader("From", "haborhuang@whispir.cc")
	msg.SetHeader("To", "haborhuang@whispir.cc")
	msg.SetHeader("Subject", "Hello world")
	msg.SetHeader("Message-ID", encodeId(mid, appid))
	msg.SetBody("text/plain", "Hello Habor! I'm Habor")
	ch <- msg

	startReceiver()
	neverstop := make(chan struct{})
	<-neverstop

}

func encodeId(mid, appid string) string {
	enc := base64.URLEncoding
	return fmt.Sprintf("<%s.%s@whispir.cn>",
		enc.EncodeToString([]byte(mid)),
		enc.EncodeToString([]byte(appid)),
	)
}

func startSender(ch chan *gomail.Message)  {
	go func() {
		d := gomail.NewDialer("smtp.exmail.qq.com", 465, user, password)

		var s gomail.SendCloser
		var err error
		open := false
		for {
			select {
			case msg := <- ch:
				if !open {
					if s, err = d.Dial(); nil != err {
						panic(err)
					}
					open = true
				}
				if err := gomail.Send(s, msg); nil != err {
					log.Println(err)
					continue
				}
				log.Println("Email was sent")
			case <-time.After(30 * time.Second):
				if open {
					if err := s.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}()
}

func startReceiver()  {
	c, err := client.DialTLS("imap.exmail.qq.com:993", nil)
	if nil != err {
		log.Fatal(err)
	}

	if err := c.Login(user, password); nil != err {
		log.Fatal(err)
	}
	defer c.Logout()

	//mailboxes := make(chan *imap.MailboxInfo, 10)
	//done := make(chan error, 1)
	//go func () {
	//	done <- c.List("", "*", mailboxes)
	//}()
	//
	//log.Println("Mailboxes:")
	//for m := range mailboxes {
	//	log.Println("* " + m.Name)
	//}
	//
	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}

	mbox, err := c.Select("INBOX", false)
	if nil != err {
		log.Fatal(err)
	}
	if mbox.Messages < 1 {
		log.Println("No messages")
		return
	}
	//
	//from := uint32(1)
	//to := mbox.Messages
	//if mbox.Messages > 3 {
	//	from = mbox.Messages - 3
	//}
	//seqNums, err := c.UidSearch(&imap.SearchCriteria{WithoutFlags:[]string{imap.SeenFlag}})
	//if nil != err {
	//	log.Fatal(err)
	//}
	//if len(seqNums) < 1 {
	//	log.Println("No Unseen messages")
	//	return
	//}
	//log.Println("searched sequences:", seqNums)
	seqset := new(imap.SeqSet)
	//seqset.AddNum(seqNums...)
	seqset.AddRange(mbox.Messages - 1, mbox.Messages - 1)
	//
	messages := make(chan *imap.Message, 1)

	go func() {
		if err := c.Fetch(seqset, []string{"BODY[]", imap.Uid, imap.FlagsMsgAttr}, messages); nil != err {
			log.Fatal(err)
		}
	}()

	for msg := range messages {
		r := msg.GetBody("BODY[]")
		log.Println(msg.Flags)
		log.Println(msg.Uid)
		if nil == r {
			log.Fatal("Server didn't returned message body")
		}

		//log.Println(ioutil.ReadAll(r))
		mr, err := mail.CreateReader(r)
		if nil != err {
			log.Fatal(err)
		}

		header := mr.Header
		log.Printf("headers: %#v\n", header)
		//log.Println("===================================")
		//subject, _ := header.Subject()
		//log.Println("Subject:", subject)
		//from, _ := mail2.ParseAddress(header.Get("From"))
		//fromName, _ := charset.DecodeHeader(from.Name)
		//log.Println("From:", fromName, from.Address)
		//to, _ := mail2.ParseAddress(header.Get("To"))
		//toName, _ := charset.DecodeHeader(to.Name)
		//log.Println("To:", toName, to.Address)
		//log.Println("In-Reply-To:", header.Get("In-Reply-To"))
		//log.Println("Message-Id:", header.Get("Message-Id"))
		//log.Println("Message-ID:", header.Get("Message-ID"))

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if nil != err {
				log.Fatal(err)
			}

			switch h := p.Header.(type) {
			case mail.TextHeader:
				b, _ := ioutil.ReadAll(p.Body)
				log.Println(h.ContentType())
				log.Printf("Got message: %v\n", string(b))
			//case mail.AttachmentHeader:
			//	filename, _ := h.Filename()
			//	log.Printf("Got attachment: %v\n", filename)
			}
		}
	}

	//if err := c.UidStore(seqset, imap.AddFlags, []interface{}{imap.SeenFlag}, nil); nil != err {
	//	log.Fatal(err)
	//}
	log.Println("Done")
}