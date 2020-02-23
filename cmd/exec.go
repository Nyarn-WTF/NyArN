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
	homedir "github.com/mitchellh/go-homedir"
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
				ErrorAlert(string(out) + "\n\nstderr output\n" + err.Error())
				log.Fatal(err)
			}
		default:
			out, err := exec.Command(c[0], c[1:]...).Output()
			fmt.Println(string(out))
			if err != nil {
				ErrorAlert(string(out) + "\n\nstderr output\n" + err.Error())
				log.Fatal(err)
			}
		}
	},
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&sound, "sound", "s", false, "Alert's sound enable flag")
	execCmd.PersistentFlags().BoolVarP(&email, "mail", "m", false, "Alert e-mail enable flag")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", home+"/.NyArN_config.yaml", "config file name")
}

//Config ファイルから読み込む構造体
type Config struct {
	Soundfile string
	Username  string
	Password  string
	Rcpt      string
	Host      string
}

//ErrorAlert アラートまとめ関数
func ErrorAlert(execError string) {
	fmt.Println(configFile)
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
		return
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
		return
	}

	if email {
		err := SendEmail(config.Username, config.Password, config.Rcpt, config.Host, execError)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if sound {
		err := Sound(config.Soundfile)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

//Sound 音を鳴らす関数
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

//SendEmail メールを送る関数
func SendEmail(username string, password string, rcpt string, host string, message string) error {
	server := host + ":" + "587"
	fmt.Println(server)
	body := message

	auth := smtp.PlainAuth("", username, password, host)

	sendMessage := "To: " + rcpt + "\r\n"
	sendMessage += "Subject: " + "NyArN Build Error" + "\r\n\r\n"
	sendMessage += body + "\r\n"
	fmt.Println(username)
	if err := smtp.SendMail(server, auth, username, []string{rcpt}, []byte(sendMessage)); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
