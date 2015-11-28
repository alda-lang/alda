// DEFAULT_PORT = 27713

// puts 'Starting Alda server on port 27713...'
// pid = spawn "cd ~/Code/alda && boot alda -x 'server --port 27713'", out: '/dev/null', err: '/dev/null'
// Process.detach pid

package main

import (
  "os"
  "github.com/codegangsta/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "alda"
  app.Version = "0.14.2"
  app.Usage = "a music language for musicians ♬ ♪"
  app.Action = func(c *cli.Context) {
    println("TODO")
  }

  app.Run(os.Args)
}
