# hexx

`hexx` is a small colorized hex dumper for inspecting files. It displays byte offsets, grouped hexadecimal bytes, and an ASCII preview. Search matches are highlighted in the hex column.

## Usage

```sh
go run . [options] <file>
```

Build an executable with:

```sh
go build -o hexx .
./hexx <file>
```

Options must come before the filename.

| Option | Description | Default |
| --- | --- | --- |
| `-c <n>` | Number of bytes displayed per row | `16` |
| `-g <n>` | Number of bytes per visual hex group | `1` |
| `-F <text>` | Search for literal text | None |
| `-f <hex>` | Search for hex bytes, written as pairs of digits | None |

## Examples

Display a file with the default layout:

```sh
go run . capture.pcap
```

Display 16 bytes per row, grouped in pairs:

```sh
go run . -c 16 -g 2 capture.pcap
```

Search for a literal string:

```sh
go run . -F "hello" file.bin
```

Search for the byte sequence `d4 c3 b2 a1`:

```sh
go run . -f d4c3b2a1 capture.pcap
```

Search for both a string and a hex sequence in the same dump:

```sh
go run . -F "hello" -f d4c3b2a1 capture.pcap
```

Literal-string matches are highlighted with a magenta background; hex matches use cyan. `-f` accepts pairs of hexadecimal digits without separators, and the number of digits must be even.