# Development Environment Setup
* Install [Golang](https://golang.org/doc/install?download=go1.12.17.linux-amd64.tar.gz)
* Dependencies are installed via go modules automatically
* Ensure GOPATH matches the repo path
    * `echo $GOPATH`
* Install the arm-linux crosscompiler (Required for SQLite):
    * `sudo apt-get install gcc-arm-linux-gnueabihf`
* Recommend using [Visual Studio Code](https://code.visualstudio.com/) as the Code Editor
* Recommend using the git terminal for building
* Install recommended Go extension from microsoft in VS Code
* Need to use import tool for formatting, add this to your Go extension settings in VSCode:
    * go.formatTool: "goimports"
* Ctrl+Shift+P in VSCode to open Command Pallete: 
    * run "Go: Install/Update Tools"

# Building
* For building just the server and host, use the script buildArmSH.sh
    * Example: `./buildArmSH.sh`
    * Outputs can be found under the `armbin/` folder
* For building apps that run on the target, use the script buildArm.sh.  
    * Example: `./buildArm.sh server`

# Debugging
To observe any logged statements produced by host or server, run them with the `-l` argument for additional information, and the `-d` argument for debugging information, and `&` to run in the background. This should be done in PuTTY when connected to a device.

* Example below:
    * `tcpHost &`


## tcpServer
This hosts the http server.  It will serve up webpages located in specific directory (see routeHandlers.go) and also has a REST interface.
* Use `./buildArm.sh tcpserver`