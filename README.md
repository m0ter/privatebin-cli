# Privatebin CLI

> Tested on ubuntu 18 and 22

Build binary

```bash
go build .
```

or download the latest built binary

https://github.com/M0ter/privatebin-cli/releases/latest/download/privatebin-linux-amd64

Create config .privatebin.yaml in $PWD or $HOME/ or use the --url flag with contents

```yaml
url: https://privatebin.net
```

expires, burn and delete can also be configured in the config

ex.
```yaml
url: https://privatebin.net
expires: 5min
burn: true
format: plain
output: rich
```

If the config has burn: true and you want to override you can do that by passing false `--burn=false`

Run the binary
```bash
testing@pc:~$ ./privatebin -h
CLI access to privatebin

Usage:
  privatebin "string for privatebin"... [flags]

Examples:
privatebin "encrypt this string" --expires 1day --burn --password Secret
cat textfile | privatebin --url https://yourprivatebin.com

Flags:
  -B, --burn              Burn after reading
      --config string     config file (default is $HOME/.privatebin.yaml or $PWD/.privatebin.yaml)
      --expires string    How long the snippet should live
                          5min, 10min, 1hour, 1day, 1week, 1month, 1year, never (default "5min")
      --format string     Paste format
                          plain, code, md (default "plaintext")
  -h, --help              help for privatebin
      --output string     Output format of the returned data
                          simple, rich, json (default "simple")
      --password string   Password for the paste
      --url string        URL to privatebin app (default "https://privatebin.net")
  -v, --verbose           verbose output
      --version           version for privatebin
```

```bash
testing@pc:~$ ./privatebin "encrypt this string" --burn
https://privatebin.net/?1a803063047a7672#289BHS9hJBPpsqVeHqG8bWk8dLpMSreAcx9QGKQb5Gox
```

```bash
testing@pc:~$ echo "hello\nworld" | ./privatebin --expires never -B --output rich
Secret URL: https://privatebin.net/?bdf5bac30d32e74b#92vEmyscPwfvyRcbZ1eRxf2mZ1HKdhWdm4Q7iWNEnU3s
Delete URL: https://privatebin.net/?pasteid=bdf5bac30d32e74b&deletetoken=2aaeee6ccd4d54e57e86483c6e3d02fcd1430078bdce2249089f62bb411bfb69
```
