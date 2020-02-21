/*
Copyright © 2020 Nyarn-WTF <s16054@tokyo.kosen-ac.jp>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sound bool
var email bool
var configFile string
var config Config

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "execute command",
	Run: func(cmd *cobra.Command, args []string) {
		//引数がなかったときの処理
		if len(args) == 0 {
			fmt.Println("This subcommand is execute argument command.")
			return
		}

		//引数を分離
		c, err := shellwords.Parse(args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(c)

		//引数の引数の数によって処理を分離
		switch len(c) {
		case 0:
			log.Fatal(err)
		case 1:
			out, err := exec.Command(c[0]).Output()
			fmt.Println(string(out))
			if err != nil {
				ErrorAlert(err)
				log.Fatal(err)
			}
		default:
			out, err := exec.Command(c[0], c[1:]...).Output()
			fmt.Println(string(out))
			if err != nil {
				ErrorAlert(err)
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&sound, "sound", "s", false, "Alert's sound enable flag")
	execCmd.PersistentFlags().BoolVarP(&email, "mail", "m", false, "Alert e-mail enable flag")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "~/.NyArN_config.yaml", "config file name")
}

type Config struct {
	soundfile string
	user      string
	password  string
	rcpt      string
	host      string
}

func ErrorAlert(execError error) {
	viper.SetConfigFile(configFile)

	if err := viper.ReadConfig(nil); err != nil {
		log.Fatal(err)
		return
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
		return
	}

	if email {
		err := SendEmail(config.user, config.password, config.rcpt, config.host, execError)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if sound {
		err := Sound(config.soundfile)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

func Sound(path string) error {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("File Open Error")
		return err
	}
	s, format, err := mp3.Decode(f)
	if err != nil {
		fmt.Println("Decode Error")
		return err
	}
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		fmt.Println("Init Sound Error")
		return err
	}
	done := make(chan struct{})
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))
	<-done

	return nil
}

func SendEmail(user string, password string, rcpt string, host string, message error) error {
	server := host + ":" + "465"
	body := message.Error()

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	auth := smtp.PlainAuth("", user, password, host)

	con, err := tls.Dial("tcp", server, tlsconfig)
	if err != nil {
		log.Fatal(err)
		return err
	}

	c, err := smtp.NewClient(con, host)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err = c.Auth(auth); err != nil {
		log.Fatal(err)
		return err
	}

	if err = c.Mail(user); err != nil {
		log.Fatal(err)
		return err
	}

	if err = c.Rcpt(rcpt); err != nil {
		log.Fatal(err)
		return err
	}

	w, err := c.Data()
	if err != nil {
		log.Fatal(err)
		return err
	}

	sendMessage := "From: " + user + "\r\n"
	sendMessage += "To: " + rcpt + "\r\n"
	sendMessage += "\n" + body

	_, err = w.Write([]byte(sendMessage))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer w.Close()

	c.Quit()
	return nil
}
