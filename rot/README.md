A simple CLI built using the [Cobra library](github.com/spf13/cobra). Cobra is used in some of the most notable Go projects including Kubernetes and Hugo.

```shell
$ go run ./main.go --help
Rot (i.e. rotate) is a Ceasar cipher tool that by default uses ROT13.

ROT13 is used in online forums as a means of hiding spoilers, punchlines, puzzle solutions, and offensive materials from the casual glance. ROT13 has inspired a variety of letter and word games on-line, and is frequently mentioned in newsgroup conversations.
    - https://en.wikipedia.org/wiki/ROT13

Usage:
  rot [string to rotate] [flags]

Flags:
  -a, --amount int   amount to rotate alphabet letters (default 13)
  -h, --help         help for rot
  -u, --unrot        undoes rotation
  ```