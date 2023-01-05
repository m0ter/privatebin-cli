# Privatebin CLI

> Tested on ubuntu 18 and 22

Install required packages for the browser:

```bash
$ sudo apt install -y libnss3 libatk-bridge2.0-0 libcups2 libgbm1 libxkbcommon-x11-0 libpango-1.0-0 libcairo2 libasound2
```

Build binary

```bash
go build .
```

Create config .privatebin.yaml in $PWD or $HOME/ or use the --url flag

```
url: https://privatebin.net
```

Run the binary
```
testing@pc:~$ ./privatebin -h
CLI access to privatebin

Usage:
  privatebin "string for privatebin"... [flags]

Examples:
privatebin "encrypt this string" --expires 5min -B --password Secret
        cat textfile | privatebin --expires 5min --url https://privatebin.net
echo "hello\nworld" | privatebin --expires never -B

Flags:
  -B, --burn              Burn after reading
      --config string     config file (default is $HOME/.privatebin.yaml or $PWD/.privatebin.yaml)
  -D, --delete            Show delete link
      --expires string    Required flag for how long the snippet should live
                          5min, 10min, 1hour, 1day, 1week, 1month, 1year, never
  -h, --help              help for privatebin
      --password string   Password for the snippet
      --url string        URL to privatebin app
  -v, --version           version for privatebin
```

```
testing@pc:~$ ./privatebin "encrypt this string" --expires 5min -B
https://privatebin.net/?1a803063047a7672#289BHS9hJBPpsqVeHqG8bWk8dLpMSreAcx9QGKQb5Gox
```

```
testing@pc:~$ echo "hello\nworld" | ./privatebin --expires never -B -D
Secret URL: https://privatebin.net/?bdf5bac30d32e74b#92vEmyscPwfvyRcbZ1eRxf2mZ1HKdhWdm4Q7iWNEnU3s
Delete URL: https://privatebin.net/?pasteid=bdf5bac30d32e74b&deletetoken=2aaeee6ccd4d54e57e86483c6e3d02fcd1430078bdce2249089f62bb411bfb69
```