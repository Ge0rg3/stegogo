# stegogo
A Go tool to perform a large number of steganographic operations.

## Examples
### LSB
* Embed `secret.txt` file within `cats.png` in the R0, B2, then R1 planes: 
```
stegogo lsb embed --secret secret.txt --cover cats.png --output cats.png R0 B2 R1
```
* Extract the embedded data from `cats.png` from R0, B2 then R1 planes:
```
stegogo lsb extract --input cats.png R0 B2 R1
```
* Embed `secret.zip` within `cats.png` in the Alpha 0 plane, column by column:
```
stegogo lsb embed --secret secret.txt --cover cats.png --output cats.png --column A0
```
