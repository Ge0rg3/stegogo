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
stegogo lsb embed --secret secret.zip --cover cats.png --output cats.png --column A0
```

### PVD
* Embed `secret.txt` file within `cats.png` greyscale image with default range widths (8 8 16 32 64 128):
```
stegogo pvd embed --secret secret.txt --cover cats.png --output cats_secret.png
```
* Extract secret data from `cats_secret.png` greyscale image with default range widths (8 8 16 32 64 128).
```
stegogo pvd extract --input cats_secret.png --output secret.dat
```
* Embed `secret.png` file within `cats.png` greyscale image in columnar order, with zigzag, with custom range widths (2 2 4 4 4 8 8 16 16 32 32 64 64):
```
stegogo pvd embed --secret secret.png --cover cats.png --output cats_secret.png --direction column --zigzag 2 2 4 4 4 8 8 16 16 32 32 64 64
```
* Extract secret data from `cats_secret.png` greyscale image in columnar order, with zigzag, with custom range widths (2 2 4 4 4 8 8 16 16 32 32 64 64):
```
stegogo pvd extract --input cats_secret.png --output secret.dat --direction column --zigzag 2 2 4 4 4 8 8 16 16 32 32 64 64
```

### Exif
* Embed `secret.txt` file within `cats.png`, in the `ProcessingSoftware` EXIF tag:
```
stegogo exif -i cats.png -s secret.txt
```
* Extract all EXIF tags from `cats.png`:
```
stegogo exif -i cats.png
```

### Bit Plane Steganography
* Embed a black and white image `bw.png` within `cats.png`, in the R0, B0 and G0 planes:
```
stegogo bp embed -c cats.png -s bw.png -o hidden.png R0 G0
```
* Extract a hidden image from `cats.png` only if there are bits in both the R0 and G0 planes:
```
stegogo bp extract -i cats.png -o hidden.png R0 G0
```

### Peak Signal-to-Noise Ratio
* Read the PSNR of `a.png` and `b.png`:
```
stegogo psnr -i a.png -c b.png
```