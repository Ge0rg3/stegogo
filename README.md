# stegogo
A Go tool to perform a large number of steganographic operations.

## Examples
### LSB
Embed `secret.txt` data within `cats.png` file in the R0 bit plane.
stegogo lsb embed --secret secret.txt --cover cats.png --output cats.png R0 B2 R1