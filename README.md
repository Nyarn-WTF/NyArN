Nya-Alert-Neko is a build alert watchdog.

# Installation
1. Install go env -> [https://golang.org/](https://golang.org/)
2. Clone this repo \
   `git clone https://github.com/Nyarn-WTF/NyArN.git`
3. build & install
   ```
   cd NyArn
   go build
   go install
   cp ./config.yaml ~/.NyArN_config.yaml
   ```
   and,edit `~/.NyArN_config.yaml`
4. Add install-path \
   In ~/.zshrc \
   `
   export PATH="$PATH:/home/[username]/go/bin"
   ` \
   then, \
   `source ~/.zshrc`
